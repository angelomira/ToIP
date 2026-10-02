New-Item -ItemType Directory -Force -Path .\bin | Out-Null
go build -o .\bin\server.exe .\cmd\server