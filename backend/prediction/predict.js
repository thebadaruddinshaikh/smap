import { exec } from 'child_process';
import path from 'path';
import { fileURLToPath } from 'url';

// Wrap the exec function in a promise to use async/await
function execPromise(command) {
    return new Promise((resolve, reject) => {
        exec(command, (error, stdout, stderr) => {
            if (error) {
                reject(`Error executing Python script: ${error}`);
            } else if (stderr) {
                reject(`Python stderr: ${stderr}`);
            } else {
                resolve(stdout);
            }
        });
    });
}

export default async function predict(imagePath) {
    const __filename = fileURLToPath(import.meta.url);
    const __dirname = path.dirname(__filename);
    const pythonScript = path.join(__dirname, 'predict.py');

    const command = `python3 ${pythonScript} ${imagePath}`;
    
    try {
        const result = await execPromise(command);
        return result;
    } catch (error) {
        console.error(error);
    }
}
