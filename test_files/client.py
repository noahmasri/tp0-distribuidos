import socket
import time

def send_messages():
    try:
        for i in range(10):
            with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
                s.connect(('127.0.0.1', 12345))
                message = f"Message {i}"
                s.sendall(message.encode('utf-8'))
                
                # Leer la respuesta del servidor
                response = s.recv(1024)
                print(f"Received from server: {response.decode('utf-8')}")
                s.close()
                time.sleep(2)

    except Exception as e:
        print(f"An error occurred: {e}")

if __name__ == "__main__":
    send_messages()

