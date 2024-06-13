from django.utils import timezone
from celery import shared_task
import subprocess
import socket
import requests
import json

from common.models import ScanJob
from .models import Subdomain  # 确保正确导入模型
from django.db import transaction

def resolve_ip(subdomain):
    try:
        return socket.gethostbyname(subdomain)
    except socket.gaierror:
        return None

def check_http_https(url):
    protocols = ['http', 'https']
    responses = {}
    for protocol in protocols:
        try:
            response = requests.get(f"{protocol}://{url}", timeout=1)
            responses[protocol] = {
                'status_code': response.status_code,
                'headers': dict(response.headers)
            }
        except requests.exceptions.RequestException as e:
            responses[protocol] = {
                'status_code': 000,  # 如果发生异常，则设置状态码为000
                'error': str(e)
            }
    return responses

@shared_task(bind=True)
def scan_subdomains(self, target, from_job_id=None):
    # 创建SubdomainScanJob实例
    scan_job = ScanJob.objects.create(
        type='SUBDOMAIN',
        target=target,
        status='R',
        task_id=self.request.id,
        from_job_id=from_job_id,
    )

    # 构建输出文件名
    output_file_path = f"/tmp/{scan_job.task_id}.json"

    # 构建Subfinder命令
    cmd = f"subfinder -d {target} -all -o {output_file_path} -oJ"

    try:
        # 执行Subfinder命令
        process = subprocess.run(cmd, shell=True, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)

        # 从输出文件读取结果
        with open(output_file_path, 'r') as file:
            for line in file:
                result = json.loads(line.strip())  # 逐行读取并解析JSON
                subdomain = result.get('host')
                source = result.get('source', '')
                ip_address = resolve_ip(subdomain)  # 解析IP地址

                # 检查子域名的HTTP和HTTPS
                subdomain_http_https_results = check_http_https(subdomain)
                # 如果有IP地址，对IP进行HTTP和HTTPS检测
                ip_http_https_results = check_http_https(ip_address) if ip_address else {}

                # 创建Subdomain对象
                with transaction.atomic():
                    Subdomain.objects.create(
                        scan_job=scan_job,
                        subdomain=subdomain,
                        domain=target,
                        source=source,
                        ip_address=ip_address,
                        subdomain_http_status=subdomain_http_https_results['http'].get('status_code', ''),
                        subdomain_https_status=subdomain_http_https_results['https'].get('status_code', ''),
                        ip_http_status=ip_http_https_results.get('http', {}).get('status_code', ''),
                        ip_https_status=ip_http_https_results.get('https', {}).get('status_code', ''),
                        from_asset=target,
                    )
            scan_job.status = 'C'  # 标记为完成
    except subprocess.CalledProcessError as e:
        scan_job.status = 'E'  # 标记为错误
        scan_job.error_message = f'子域名扫描失败: {e.stderr.decode()}'
    except Exception as e:
        scan_job.status = 'E'  # 标记为错误
        scan_job.error_message = f'处理子域名扫描结果时发生异常: {str(e)}'
    finally:
        scan_job.end_time = timezone.now()
        scan_job.save()

        if scan_job.status == 'E':
            return {'error': scan_job.error_message}
        return {'message': f'子域名扫描完成: {target}'}