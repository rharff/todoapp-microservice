@Library('jenkins-pipeline-library') _

standardDevopsMicroservice(
    appName: "todoapp",
    services: [
        [
            name: "Task Service",
            path: "services/task-service",
            image: "devopsnaratel/todoapp-task",
            buildCmd: "go test ./... && go build -o bin/server ./cmd/server"
        ],
        [
            name: "Audit Service",
            path: "services/audit-service",
            image: "devopsnaratel/todoapp-audit",
            buildCmd: "go test ./... && go build -o bin/server ./cmd/server"
        ],
        [
            name: "Web Frontend",
            path: "frontend",
            image: "devopsnaratel/todoapp-ui",
            buildCmd: "npm install && npm run build"
        ]
    ]
)