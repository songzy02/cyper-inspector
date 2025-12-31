package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os/exec"

	"github.com/gin-gonic/gin"
)

type AgentResponse struct {
	Hostname string          `json:"hostname"`
	RawData  json.RawMessage `json:"raw_data"`
	Analysis json.RawMessage `json:"analysis"`
}

func inspectHandler(c *gin.Context) {
	cmd := exec.Command("bash", "-c", `
hostname=$(hostname); timestamp=$(date -Iseconds)
cpu_used=$(top -bn1 | awk '/Cpu/{print 100-$8"%"}')
cpu_load=$(uptime | awk -F'load average:' '{print $2}' | awk '{print $1}' | sed 's/,//')
mem_used=$(free -m | awk 'NR==2{printf "%.2f%%",$3/$2*100}')
disk_alert=$(df -P -x tmpfs -x devtmpfs -x overlay -x shm | \
             awk '$1 !~ /loop/ && $6 !~ /^\/var\/lib\/docker/ && $6 !~ /^\/kubelet/ && $5+0>80{
                    gsub(/%/,"",$5); print $1":"$6":"$5"%"
                  }' | paste -sd';')
raid_state=$(/opt/MegaRAID/MegaCli/MegaCli64 -LDInfo -Lall -aALL 2>/dev/null |
             grep -E "State\s*:" | awk '{print $3}' | paste -sd, -)
[ -z "$raid_state" ] && raid_state="null"
err_cnt=$(journalctl -p err --since "1 hour ago" 2>/dev/null | wc -l)
gateway=$(ip route | awk '$1=="default"{print $3}')
loss=$(ping -c 10 -W 1 "$gateway" 2>/dev/null | grep -o '[0-9]*%' | tr -d '%')
jq -c -n \
  --arg hn "$hostname" \
  --arg ts "$timestamp" \
  --arg cpu "$cpu_used" \
  --arg load "$cpu_load" \
  --arg mem "$mem_used" \
  --arg disk "$disk_alert" \
  --arg raid "$raid_state" \
  --argjson err "$err_cnt" \
  --arg loss "$loss" \
  '{hostname:$hn,timestamp:$ts,cpu_used:$cpu,cpu_load:$load,mem_used:$mem,disk_alert:$disk,raid_state:$raid,journal_err_1h:$err,ping_loss:$loss}'
`)

	jsonRaw, err := cmd.Output()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据采集失败: " + err.Error()})
		return
	}

	reqBody, _ := json.Marshal(map[string]interface{}{
		"model":       "sinollm",
		"temperature": 0,
		"messages": []map[string]string{
			{"role": "system", "content": `你是一名 Linux 运维专家，只返回 JSON。阈值规则：
1. CPU 使用率 > 85% 或 1-min load > 物理核数×1.5 → CRITICAL 
2. 内存使用率 > 90% → CRITICAL 
3. 任一磁盘使用率 > 90% → CRITICAL 
4. RAID 状态 != "Optimal" → CRITICAL 
5. 1 小时内 journal 错误 > 10 条 → WARNING 
6. ping 网关丢包率 > 5% → WARNING 
输出格式：{"alert":true/false,"level":"CRITICAL|WARNING|OK","summary":"结论","details":["原因"]},"Plan":"方案"`},
			{"role": "user", "content": string(jsonRaw)},
		},
	})

	resp, err := http.Post("http://172.16.10.226:18000/v1/chat/completions",
		"application/json", bytes.NewReader(reqBody))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "LLM 请求失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	var llmResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&llmResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "LLM 响应解析失败: " + err.Error()})
		return
	}

	// 主机名写死
	hostname := "本地机器-226"

	// 安全取出 LLM 返回的 content 并转成 json.RawMessage
	choices, ok := llmResp["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "LLM 返回格式异常"})
		return
	}
	msg, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "LLM 返回格式异常"})
		return
	}
	contentStr, ok := msg["content"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "LLM 返回格式异常"})
		return
	}

	c.JSON(http.StatusOK, AgentResponse{
		Hostname: hostname,
		RawData:  jsonRaw,
		Analysis: json.RawMessage(contentStr),
	})
}

func main() {
	r := gin.Default()
	r.GET("/inspect", inspectHandler)
	_ = r.Run(":8083") // 监听 0.0.0.0:8080
}
