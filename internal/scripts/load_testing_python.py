import socket
import threading
import random
import string
import time

# Server address and port
HOST = 'localhost'
PORT = 9090

# Number of concurrent connections
NUM_CONNECTIONS = 100

# Function to generate a random key and value
def generate_random_key_value():
    key = ''.join(random.choices(string.ascii_uppercase + string.digits, k=8))
    value = ''.join(random.choices(string.ascii_uppercase + string.digits, k=12))
    return key, value

# Function to handle each connection
def handle_connection(connection_id):
    try:
        # Create a TCP/IP socket
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            # Connect to the server
            s.connect((HOST, PORT))

            # Generate random key and value
            key, value = generate_random_key_value()

            # Send the SET command to the server
            command = f"SET {key} {value}\n"
            s.sendall(command.encode())

            # Receive the response from the server
            response = s.recv(1024).decode()
            print(f"Connection {connection_id} received: {response}")

    except Exception as e:
        print(f"Connection {connection_id} encountered an error: {e}")

# Main function to start the load test
def main():
    threads = []

    # Create threads for concurrent connections
    for i in range(NUM_CONNECTIONS):
        thread = threading.Thread(target=handle_connection, args=(i,))
        threads.append(thread)
        thread.start()

    # Wait for all threads to complete
    for thread in threads:
        thread.join()

if __name__ == "__main__":
    start_time = time.time()
    main()
    print(f"Load test completed in {time.time() - start_time} seconds")
