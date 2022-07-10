job "demo" {

  datacenters = ["dc1"]
  type = "service"

  update {
    max_parallel = 1
    min_healthy_time = "10s"
    healthy_deadline = "3m"
    progress_deadline = "10m"
    auto_revert = false
    canary = 0
  }

  migrate {
    max_parallel = 1
    health_check = "checks"
    min_healthy_time = "10s"
    healthy_deadline = "5m"
  }
  group "app" {

    count = 1

    network {
      port "http" {
        to = 8080
      }

      port "prometheus" {
        to = 8090
      }
    }

    service {
      name = "demo-app-http"
      tags = ["global", "public"]
      port = "http"
    }

    service {
      name = "demo-app-prometheus"
      tags = ["prometheus"]
      port = "prometheus"
    }

    restart {
      attempts = 2
      interval = "30m"
      delay = "15s"
      mode = "fail"
    }

    task "demo-app" {
      driver = "docker"

      config {
        image = "wfaler/demo-app:v15"
        ports = ["http", "prometheus"]

      }
      // env {
      //   TEMPO_ENDPOINT = "10.0.0.7:4317"
      // }

      template {
        data = <<EOH
# Lines starting with a # are ignored
{{- range service "tempo-grpc" }}
TEMPO_ENDPOINT="{{ .Address }}:{{ .Port }}"
{{- end }}
FOO=bar
      EOH
//this is how you get consul kv and vault secrets
#LOG_LEVEL="{{key "service/geo-api/log-verbosity"}}"
#API_KEY="{{with secret "secret/geo-api-key"}}{{.Data.value}}{{end}}"

        env         = true
        destination = "/app/env"
      }
  }
}
