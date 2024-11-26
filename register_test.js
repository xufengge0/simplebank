import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
    stages: [
        { duration: '40s', target: 200 }, // 在 30 秒内增加到 1000 个并发用户
        { duration: '10s', target: 200 },  // 持续 30 秒保持 1000 个并发用户
        { duration: '40s', target: 0 },   // 30 秒内将并发用户减少到 0
    ],
};
// 生成指定长度的随机小写字母字符串
function getRandomString(length) {
    let result = '';
    let characters = 'abcdefghijklmnopqrstuvwxyz';
    let charactersLength = characters.length;
    for (let i = 0; i < length; i++) {
        result += characters.charAt(Math.floor(Math.random() * charactersLength));
    }
    return result;
}

export default function () {
    let url = 'http://localhost:8080/v1/create_user';

    // 生成6个小写字母的随机用户名和邮箱名
    let randomUsername = getRandomString(16);
    let randomEmail = getRandomString(16);

    let payload = JSON.stringify({
        username: `${randomUsername}`,  // 动态生成用户名，防止重复
        password: 'securepassword',
        full_name: `${randomUsername}`,
        email: `${randomEmail}@example.com`
    });

    let params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    // 阻塞的请求
    let res = http.post(url, payload, params);

    // 检查响应是否为 201 Created 非阻塞
    check(res, { 'status was 200': (r) => r.status === 200 });

    // 模拟用户思考时间
    sleep(1);
}
