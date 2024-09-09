const net = require('net');
const crypto = require('crypto');

// TCP server address
const address = 'localhost';
const port = 9090;
const numConnections = 150; // Number of concurrent connections
const keySize = 8; // Size of the random key
const valueSize = 12; // Size of the random value

// Function to generate random strings for key and value
function randomString(length) {
    return crypto.randomBytes(length).toString('hex').slice(0, length);
}

// Function to handle individual connection
function handleConnection(connID) {
    return new Promise((resolve) => {
        const client = new net.Socket();

        client.connect(port, address, () => {
            const key = randomString(keySize);
            const value = randomString(valueSize);

            // Send SET command
            const command = `SET ${key} ${value}\n`;
            client.write(command);
        });

        client.on('data', (data) => {
            console.log(`Connection ${connID} received: ${data.toString().trim()}`);
            client.destroy(); // Close the connection after receiving the response
            resolve();
        });

        client.on('error', (err) => {
            console.log(`Connection ${connID} error: ${err.message}`);
            resolve();
        });

        client.on('close', () => {
            resolve();
        });
    });
}

async function runLoadTest() {
    const startTime = Date.now();
    const promises = [];

    // Launch concurrent connections
    for (let i = 0; i < numConnections; i++) {
        promises.push(handleConnection(i));
    }

    // Wait for all connections to complete
    await Promise.all(promises);

    // Print total time taken for the load test
    const duration = (Date.now() - startTime) / 1000;
    console.log(`Load test completed in ${duration} seconds`);
}

runLoadTest();