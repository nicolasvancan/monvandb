import socket
import ssl

def main():
    host = "localhost"
    port = 8080

    # Configure SSL context
    context = ssl.SSLContext(ssl.PROTOCOL_TLS_CLIENT)
    # For self-signed certificates, disable hostname and certificate verification
    context.check_hostname = False
    context.verify_mode = ssl.CERT_NONE
    # For production with CA-issued certificates, enable verification:
    # context.verify_mode = ssl.CERT_REQUIRED
    # context.load_verify_locations("server.crt")  # Path to server certificate or CA

    # Connect to server
    with socket.create_connection((host, port)) as sock:
        with context.wrap_socket(sock, server_hostname=host) as ssock:
            # Send credentials
            credentials = "mvdb://admin:password@system"
            ssock.sendall(credentials.encode())

            # Receive welcome message
            data = ssock.recv(1024)
            if not data:
                print("No response from server")
                return
            print("Received:", data.decode())

            # Send a sample message
            ssock.send(b"Hello, server!")
            print("Sent: Hello, server!")

            message = input("Write something to be sent to the server: ")

            while message != "exit":
                ssock.send(message.encode())
                print(f"Sent: {message}")
                data = ssock.recv(1024)
                if not data:
                    print("No response from server")
                    break
                print("Received:", data.decode())
                message = input("Write something to be sent to the server: ")

if __name__ == "__main__":
    main()