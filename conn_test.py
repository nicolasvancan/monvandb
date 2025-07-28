import socket
import ssl
import pandas as pd
MONVAN_EOF = "d2a81ac4c97aa94e0c3474c619cc85508a19544f388c9c1c0d26773a2b19eb09"
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
    import json
    # Connect to server
    with socket.create_connection((host, port)) as sock:
        with context.wrap_socket(sock, server_hostname=host) as ssock:
            # Send credentials
            credentials = {"type": "authenticate", "message":"mvdb://admin:password@system"}
            ssock.sendall((json.dumps(credentials)+MONVAN_EOF).encode())

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
                ssock.send((message + MONVAN_EOF).encode())
                print(f"Sent: {message}")
                data = ssock.recv(1024)
                if not data:
                    print("No response from server")
                    break
                print("Received:", data.decode())
                message = input("Write something to be sent to the server: ")

from client.monvan_client import MonvanConnection, connect

if __name__ == "__main__":
    conn: MonvanConnection = connect("mvdb://admin:password@localhost:8080/my_test_db_23")
    # sleep for one second to ensure connection is established
    pandas_df: pd.DataFrame = conn.query("SELECT * FROM teste", as_pandas=True)
    print(pandas_df)