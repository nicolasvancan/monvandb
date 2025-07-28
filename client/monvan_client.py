"""
Monvan Client Module
This module provides a client for connecting to the Monvan server.
It includes methods for connecting to the server, sending messages, and receiving responses.
"""
from typing import Optional
import socket
import ssl
import json
import pandas as pd

MONVAN_EOF = "d2a81ac4c97aa94e0c3474c619cc85508a19544f388c9c1c0d26773a2b19eb09"

typesConversion = {
    0: str,
    1: int,
    2: float,
    3: bool,
    4: pd.Timestamp,
    5: str,
}

def __version__() -> str:
    """
    Returns the version of the Monvan client.
    """
    return "1.0.0"


def connect(conn_string: str, **kwargs) -> 'MonvanConnection':
    """
    Connects to the Monvan server using the provided connection string.

    :param conn_string: The connection string for the Monvan server.
    "mvdb://username:password@host:port/database"

    :param kwargs: Additional keyword arguments for connection options.

    kwargs can include:
        - timeout: Connection timeout in seconds.
        - retries: Number of retries for connection attempts.

    
    :return: An instance of MonvanConnection.
    """

    if not conn_string.startswith("mvdb://"):
        raise ValueError("Connection string must start with 'mvdb://'")
    
    parts = conn_string[7:].split('@')
    if len(parts) != 2:
        raise ValueError("Invalid connection string format. Expected 'mvdb://username:password@host:port/database'")
    
    user_info, db_info = parts
    username, password = user_info.split(':')
    host, portdatabase = db_info.split(':')
    port, database = portdatabase.split('/')

    if not host or not port or not database:
        raise ValueError("Invalid connection string format. Ensure host, port, and database are specified.")
    if not username or not password:
        raise ValueError("Username and password must be provided in the connection string.")
    
    port = int(port)

    conn = MonvanConnection(host, port, database, username, password)
    conn._connect()  # Establish the connection
    return conn

class MonvanConnection:
    def __init__(self, host: str, port: int, database: str, username: Optional[str] = None, password: Optional[str] = None):
        self.host = host
        self.port = port
        self.database = database
        self.username = username
        self.password = password
        
        self.connected = False
        self.ssl_sock = None

    def __str__(self) -> str:
        return f"MonvanConnection(host={self.host}, port={self.port}, database={self.database}, username={self.username}, connected={self.connected})"

    def _connect(self):
        sock = socket.create_connection((self.host, self.port))
        context = ssl.SSLContext(ssl.PROTOCOL_TLS_CLIENT)
        context.check_hostname = False  # Disable hostname verification for self-signed certs
        context.verify_mode = ssl.CERT_NONE  # Disable certificate verification for self-signed cert
        self.ssl_sock = context.wrap_socket(sock, server_hostname=self.host)
        self.connected = True
        message_authenticate = {
            "type": "authenticate",
            "message": f"mvdb://{self.username}:{self.password}@{self.database}"
        }
        self.__send_message(json.dumps(message_authenticate))
        rec = self.__await_answer()
        res = json.loads(rec)
        
        if res.get("type") != "success":
            raise RuntimeError(f"Connection failed: {res.get('message', 'Unknown error')}")


        print(f"Connected to {self.host}:{self.port} Database {self.database} as {self.username}")

    def __send_message(self, message: str):
        # Placeholder for sending a message
        if not self.ssl_sock:
            raise RuntimeError("Not connected to the server.")
        
        message += MONVAN_EOF
        print(f"Sending message: {message}")
        self.ssl_sock.send(message.encode())
    
    def __await_answer(self):
        if not self.ssl_sock:
            raise RuntimeError("Not connected to the server.")
        
        data = self.ssl_sock.recv(1024)
        while not data.endswith(MONVAN_EOF.encode()):
            more_data = self.ssl_sock.recv(1024)
            if not more_data:
                break
            data += more_data
        if not data:
            raise RuntimeError("No response from server.")
        
        
        return data.decode()[:-len(MONVAN_EOF)]
    
    def query(self, query: str, **kwargs) -> dict | pd.DataFrame:
        """
        Sends a query to the Monvan server and returns the response.

        :param query: The query string to send to the server.
        :kwargs: Additional keyword arguments.
            - as_pandas: If True, returns the response as a pandas DataFrame.
        :return: The response from the server as a dictionary.
        """
        if not self.connected:
            raise RuntimeError("Not connected to the server.")
        
        mes = {
            "type": "query",
            "message": query
        }

        self.__send_message(json.dumps(mes))
        response = self.__await_answer()
        if kwargs.get("as_pandas", False):
            try:
                data_json: dict = json.loads(response)
                data = data_json.get("result", {})

                if not data:
                    raise ValueError("No result returned")
                
                columns = data.get("columns", [])
                if not columns:
                    raise ValueError("No columns found in the response.")
                values: dict = data.get("data", {})
                if not values:
                    raise ValueError("No data found in the response.")
                dtypes = [typesConversion[t] for t in data.get("types", [])]
                if len(dtypes) != len(columns):
                    raise ValueError("Mismatch between number of columns and types in the response.")

                df = pd.DataFrame()
                # Create a dataframe for each column converting types avoiding int conversion issues
                for i, col in enumerate(columns):
                    if col in values.keys():
                        try:
                            df[col] = pd.Series(values[col], dtype=dtypes[i])
                        except ValueError as e:
                            if dtypes[i] == int:
                                df[col] = pd.Series(values[col], dtype=float)
                            else:
                                raise ValueError(f"Failed to convert column '{col}' to type {dtypes[i]}: {e}")
                    else:
                        raise ValueError(f"Column '{col}' not found in the response data.")
                return df
            except json.JSONDecodeError as e:
                raise ValueError(f"Failed to decode JSON response: {e}")
            
        return json.loads(response)


def build_query_message(query: str) -> str:
    """
    Builds a query message to be sent to the Monvan server.

    :param query: The query string to be sent.
    :return: The formatted query message.
    """
    if not query:
        raise ValueError("Query cannot be empty.")
    
    return json.dumps({"type":"query", "message":query}) + MONVAN_EOF

def build_disconnect_message() -> str:
    """
    Builds a disconnect message to be sent to the Monvan server.

    :return: The formatted disconnect message.
    """
    return json.dumps({"type": "close_conn"}) + MONVAN_EOF