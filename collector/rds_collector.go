package collector

import (
    "bytes"
    "log"
    "os/exec"
    "strings"

    "github.com/prometheus/client_golang/prometheus"
)


type RDSSessionCollector struct {
    sessionsDesc *prometheus.Desc
}


func NewRDSSessionCollector() *RDSSessionCollector {
    return &RDSSessionCollector{
        sessionsDesc: prometheus.NewDesc(
            "rds_sessions_total",
            "Number of RDS sessions per host and state",
            []string{"host", "state"},
            nil,
        ),
    }
}


func (c *RDSSessionCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- c.sessionsDesc
}


func (c *RDSSessionCollector) Collect(ch chan<- prometheus.Metric) {
    sessions, err := getRDSSessions()
    if err != nil {
        log.Println("Error collecting RDS sessions:", err)
        return
    }

    countMap := make(map[string]map[string]int)
    for _, s := range sessions {
        host := s.HostServer
        state := s.SessionState
        if countMap[host] == nil {
            countMap[host] = make(map[string]int)
        }
        countMap[host][state]++
    }

    for host, states := range countMap {
        for state, count := range states {
            ch <- prometheus.MustNewConstMetric(
                c.sessionsDesc,
                prometheus.GaugeValue,
                float64(count),
                host,
                state,
            )
        }
    }
}


type RDSSession struct {
    HostServer   string
    UserName     string
    SessionState string
}


func getRDSSessions() ([]RDSSession, error) {
    psCommand := `Get-RDUserSession | Select-Object HostServer,Username,SessionState`
    cmd := exec.Command("powershell", "-Command", psCommand) // could have issue with some windows languages (no issue with french/english so far)
    var out bytes.Buffer
    cmd.Stdout = &out
    err := cmd.Run()
    if err != nil {
        return nil, err
    }

    lines := strings.Split(out.String(), "\n")
    var sessions []RDSSession

    headerIndex := -1
    for i, line := range lines {
        if strings.Contains(line, "HostServer") && strings.Contains(line, "UserName") {
            headerIndex = i
            break
        }
    }
    if headerIndex == -1 || headerIndex+1 >= len(lines) {
        return sessions, nil
    }

    for _, line := range lines[headerIndex+2:] {
        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }

        // Split by whitespace
        fields := splitColumns(line)
        if len(fields) < 3 {
            continue
        }

        sessions = append(sessions, RDSSession{
            HostServer:   fields[0],
            UserName:     fields[1],
            SessionState: fields[2],
        })
    }

    return sessions, nil
}


func splitColumns(line string) []string {
    fields := strings.Fields(line)
    return fields
}
