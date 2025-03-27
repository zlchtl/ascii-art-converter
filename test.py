import requests
import json

url = "http://localhost:8080/convert"

def test_convert(image_path, size, charset):
    with open(image_path, 'rb') as img_file:
        files = {
            'file': img_file,
            'params': (None, json.dumps({'size': size, 'charSet': charset}), 'application/json')  # Параметры
        }

        response = requests.post(url, files=files)

        print(f"Testing {image_path} with size {size} and charset '{charset}'")
        print(f"Status Code: {response.status_code}")
        ascii = response.json()["ascii"]
        print(len(ascii.split("\n")[0]))
        print(ascii)
        print('-' * 50)
        return ascii

def save(s):
    with open("temp.txt","w") as f:
        f.write(s)

test_convert('images.jpg', 300, "@%#*+=-:. ")
save(test_convert('images.png', 300, "@%#*+=-:. "))