# urlShorter

Требуется создать микросервис сокращения url. Длина сокращенного URL-адреса должна быть как
можно короче. Сокращенный URL может содержать цифры (0-9) и буквы (a-z, A-Z). <br />
Эндпоинты: <br />
POST http://localhost:8080/ <br />
Request: (body): http://cjdr17afeihmk.biz/123/kdni9/z9d112423421 <br />
Response: http://localhost:8080/qtj5opu <br />
<br />
GET <br />
Request (url query): http://localhost:8080/qtj5opu <br />
Response (body): http://cjdr17afeihmk.biz/123/kdni9/z9d112423421 <br />
<br /> 
Микросервис должен уметь хранить информацию в памяти и в postgres в зависимости от флага
запуска -d
