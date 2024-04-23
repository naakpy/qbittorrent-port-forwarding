package main

import (
    "bytes"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "os/exec"
    "strconv"
    "strings"
)

func getNewPort() (int, error) {
    cmd := exec.Command("docker", "exec", "dietpi-gluetun-1", "cat", "/tmp/gluetun/forwarded_port")

    var out bytes.Buffer
    cmd.Stdout = &out
    err := cmd.Run()
    if err != nil {
        fmt.Println("Error retrieving the port:", err)
        return 0, err
    }

    forwardedPortStr := strings.TrimSpace(out.String())
    forwardedPort, err := strconv.Atoi(forwardedPortStr)
    if err != nil {
        fmt.Println("Error parsing port:", err)
        return 0, err
    }

    return forwardedPort, nil
}

func main() {
    client := &http.Client{}
    baseURL := "http://mediaserver:8090"

    newPort, err := getNewPort()
    if err != nil {
        fmt.Println("Error getting new port:", err)
        return
    }

    data := url.Values{}
    data.Set("json", fmt.Sprintf(`{"listen_port":%d}`, newPort))

    req, err := http.NewRequest("POST", baseURL+"/api/v2/app/setPreferences", strings.NewReader(data.Encode()))
    if err != nil {
        fmt.Println("Error creating request:", err)
        return
    }
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    resp, err := client.Do(req)
    if err != nil {
        fmt.Println("Error sending request:", err)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        fmt.Printf("Error: received status code %d from server\n", resp.StatusCode)
        responseBytes, err := io.ReadAll(resp.Body)
        if err == nil {
            fmt.Printf("Response: %s\n", string(responseBytes))
        }
        return
    }

    fmt.Printf("Successfully changed the listening port to %d\n", newPort)
}