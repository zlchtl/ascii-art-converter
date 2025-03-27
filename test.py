import os

#скрипт для автотестирования эндпоинтов
print("\n\n>>> curl.exe -X GET http://127.0.0.1:8080/about")
os.system("curl.exe -X GET http://127.0.0.1:8080/about")

print("\n\n>>> curl.exe -X POST http://localhost:8080/convert -H \"Content-Type: application/octet-stream\" --data-binary \"@images.jpg\"")
os.system("curl.exe -X POST http://localhost:8080/convert -H \"Content-Type: application/octet-stream\" --data-binary \"@images.jpg\"")