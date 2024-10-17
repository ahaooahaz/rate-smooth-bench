const http = require('http');

// 创建HTTP服务器
const server = http.createServer((req, res) => {
    // 处理根目录请求，返回 "Hello World"
    if (req.url === '/') {
        res.writeHead(200, { 'Content-Type': 'text/plain' });
        res.end('Hello World');
    }

    // 处理 /events 请求，逐字母推送 "Hello World"
    else if (req.url === '/events') {
        res.writeHead(200, {
            'Content-Type': 'text/event-stream',
            'Cache-Control': 'no-cache',
            'Connection': 'keep-alive',
        });

        const message = "Hello World";
        let index = 0;

        // 每0.1秒推送一个字母
        const intervalId = setInterval(() => {
            if (index < message.length) {
                res.write(`data: ${message[index++]}\n\n`);
            } else {
                // 发送结束后清除定时器
                clearInterval(intervalId);
                res.end();
            }
        }, 100); // 每100毫秒发送一次
    }

    // 处理404情况
    else {
        res.writeHead(404, { 'Content-Type': 'text/plain' });
        res.end('Not Found');
    }
});

// 设置服务器运行并60秒后关闭
server.listen(8000, () => {
    console.log('Server is running on port 8000');
    
    setTimeout(() => {
        console.log('Shutting down after 60 seconds');
        server.close();
    }, 60000);  // 60秒后关闭服务器
});
