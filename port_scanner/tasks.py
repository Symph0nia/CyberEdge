import requests
from django.utils import timezone
from celery import shared_task
from .models import ScanJob, Port
from bs4 import BeautifulSoup

import subprocess
import re
import os

@shared_task(bind=True)
def scan_ports(self, target, ports):
    # 首先创建一个新的ScanJob实例，初始化状态为'R'（Running），并立即保存
    scan_job = ScanJob.objects.create(target=target, status='R', task_id=self.request.id)

    # 创建临时文件名
    temp_file_path = f"/tmp/{scan_job.task_id}.txt"

    try:
        # 构建nmap命令，并将输出重定向到临时文件
        cmd = f"nmap -n --resolve-all -Pn --min-hostgroup 64 --max-retries 0 --host-timeout 10m --script-timeout 3m --version-intensity 9 --min-rate 10000 -T4 {target} -p {ports} -oN {temp_file_path}"
        # 执行命令
        process = subprocess.Popen(cmd, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        _, stderr = process.communicate()

        if stderr or process.returncode != 0:
            scan_job.status = 'E'  # 更新状态为错误
            scan_job.error_message = f'Nmap 扫描失败: {stderr.decode()}'
        else:
            # 从临时文件中读取输出
            with open(temp_file_path, 'r') as file:
                output = file.read()

            # 使用正则表达式查找开放的端口
            open_ports = re.findall(r'(\d+)/tcp\s+open\s+(\S+)', output)
            if not open_ports:
                scan_job.status = 'E'
                scan_job.error_message = f'没有找到开放的端口 {target}。'
            else:
                for port, service in open_ports:
                    # 创建Port对象
                    new_port = Port.objects.create(
                        scan_job=scan_job,
                        port_number=int(port),
                        service_name=service,
                        protocol='tcp',
                        state='open',
                        ip_address=target  # 添加IP地址
                    )
                    # 测试HTTP
                    new_port.http_code, new_port.http_title = check_protocol(target, port, 'http')
                    # 测试HTTPS
                    new_port.https_code, new_port.https_title = check_protocol(target, port, 'https')

                    new_port.save()

                scan_job.status = 'C'  # 更新状态为完成
    except Exception as e:
        scan_job.status = 'E'  # 更新状态为错误
        scan_job.error_message = f'扫描过程中发生异常: {str(e)}'
    finally:
        scan_job.end_time = timezone.now()  # 记录结束时间
        scan_job.save()  # 明确保存ScanJob实例的更改
        # 清理临时文件
        os.remove(temp_file_path)

        if scan_job.status == 'E':
            return {'error': scan_job.error_message}
        return {'message': f'扫描完成: {target}'}

def check_protocol(ip, port, protocol):
    url = f"{protocol}://{ip}:{port}"
    try:
        headers = {
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.71 Safari/537.36'
        }
        response = requests.get(url, headers=headers, timeout=10, verify=False)  # 禁用SSL证书验证
        status_code = response.status_code
        try:
            encoding = response.apparent_encoding
            # 使用指定的编码格式解码响应内容
            decoded_content = response.content.decode(encoding)
            # 使用BeautifulSoup解析HTML
            soup = BeautifulSoup(decoded_content, 'html.parser')
            # 获取标题
            title = soup.title.string
        except:
            title = '标题获取失败'
        # 只要请求没有引发异常，我们就认为端口支持HTTP/HTTPS
        return status_code, title
    except requests.exceptions.ConnectionError:
        # 连接错误意味着无法建立TCP连接
        return '000', '标题获取失败'
    except requests.exceptions.Timeout:
        # 超时意味着服务器没有在预定时间内响应
        return '000', '标题获取失败'
    except requests.exceptions.RequestException:
        # 处理其他所有请求相关的异常
        return '000', '标题获取失败'
