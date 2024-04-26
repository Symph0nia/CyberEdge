from django.utils import timezone
from celery import shared_task, group
import subprocess
import json
import requests
import os
import re
from subdomain_scanner.models import SubdomainScanJob, Subdomain  # 确保正确导入模型
from port_scanner.models import ScanJob, Port
from path_scanner.models import PathScanJob, PathScanResult  # 确保导入模型
from django.db import transaction

@shared_task(bind=True)
def full_scan_subdomains(self, target):
    # 创建SubdomainScanJob实例
    scan_job = SubdomainScanJob.objects.create(target=target, status='R', task_id=self.request.id)

    # 构建输出文件名
    output_file_path = f"/tmp/{scan_job.task_id}.json"

    # 构建OneForAll命令
    cmd = f"oneforall.py --target {target} --fmt json --path {output_file_path} run"

    try:
        # 执行OneForAll命令
        process = subprocess.run(cmd, shell=True, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)

        # 从输出文件读取结果
        with open(output_file_path, 'r') as file:
            results = json.load(file)
            ips = set()  # 用于存储去重的 IP 地址
            for result in results:
                # 分割和处理包含多个IP的情况
                ip_addresses = result.get('ip', '').split(',')
                for ip in ip_addresses:
                    ip = ip.strip()  # 清除空格
                    if ip:
                        ips.add(ip)
                        # 在事务中创建Subdomain对象
                        with transaction.atomic():
                            Subdomain.objects.create(
                                scan_job=scan_job,
                                subdomain=result['subdomain'],
                                ip_address=ip,
                                status=result.get('status', ''),
                                cname=result.get('cname', ''),
                                port=result.get('port', None),
                                title=result.get('title', ''),
                                banner=result.get('banner', ''),
                                asn=result.get('asn', ''),
                                org=result.get('org', ''),
                                addr=result.get('addr', ''),
                                isp=result.get('isp', ''),
                                source=result.get('source', ''),
                            )
            scan_job.status = 'C'  # 标记为完成
        # 批量启动端口扫描任务
        port_scan_tasks = group(full_scan_ports.s(ip, '1-10000') for ip in ips)
        port_scan_tasks.apply_async()

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
        return {'message': f'子域名扫描完成: {target}', 'task_id': scan_job.task_id}

@shared_task(bind=True)
def full_scan_ports(self, target, ports='1-65535'):
    scan_job = ScanJob.objects.create(target=target, status='R', task_id=self.request.id)
    temp_file_path = f"/tmp/{scan_job.task_id}.txt"

    try:
        cmd = f"nmap -sS {target} -p {ports} -oN {temp_file_path}"
        process = subprocess.Popen(cmd, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        _, stderr = process.communicate()

        if stderr or process.returncode != 0:
            scan_job.status = 'E'
            scan_job.error_message = f'Nmap 扫描失败: {stderr.decode()}'
        else:
            with open(temp_file_path, 'r') as file:
                output = file.read()
            open_ports = re.findall(r'(\d+)/tcp\s+open\s+(\S+)', output)
            if not open_ports:
                scan_job.status = 'E'
                scan_job.error_message = f'没有找到开放的端口 {target}。'
            else:
                tasks = []
                for port, service in open_ports:
                    new_port = Port.objects.create(
                        scan_job=scan_job,
                        port_number=int(port),
                        service_name=service,
                        protocol='tcp',
                        state='open',
                        ip_address=target
                    )
                    http = check_protocol(target, port, 'http')
                    https = check_protocol(target, port, 'https')
                    new_port.save()
                    url_protocol = 'https' if https else 'http'
                    url = f"{url_protocol}://{target}:{port}/FUZZ"
                    tasks.append(full_scan_paths.s('./wordlist/top7000.txt', url))

                if tasks:
                    # 启动所有路径扫描任务
                    group(tasks).apply_async()

                scan_job.status = 'C'
    except Exception as e:
        scan_job.status = 'E'
        scan_job.error_message = f'扫描过程中发生异常: {str(e)}'
    finally:
        scan_job.end_time = timezone.now()
        scan_job.save()
        os.remove(temp_file_path)

        if scan_job.status == 'E':
            return {'error': scan_job.error_message}
        return {'message': f'端口扫描完成: {target}', 'task_id': scan_job.task_id}

def check_protocol(ip, port, protocol):
    url = f"{protocol}://{ip}:{port}"
    try:
        response = requests.get(url, timeout=1, verify=False)  # 禁用SSL证书验证
        # 只要请求没有引发异常，我们就认为端口支持HTTP/HTTPS
        return True
    except requests.exceptions.ConnectionError:
        # 连接错误意味着无法建立TCP连接
        return False
    except requests.exceptions.Timeout:
        # 超时意味着服务器没有在预定时间内响应
        return False
    except requests.exceptions.RequestException:
        # 处理其他所有请求相关的异常
        return False

@shared_task(bind=True)
def full_scan_paths(self, wordlist, url):
    # 创建PathScanJob实例
    scan_job = PathScanJob.objects.create(target=url, status='R', task_id=self.request.id)

    # 构建输出文件名
    output_file_path = f"/tmp/{scan_job.task_id}.json"

    # 构建ffuf命令
    cmd = f"ffuf -w {wordlist} -u {url} -o {output_file_path} -of json"

    try:
        # 执行ffuf命令
        process = subprocess.run(cmd, shell=True, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)

        # 解析ffuf输出结果
        with open(output_file_path, 'r') as file:
            results = json.load(file)
            if 'results' in results:
                for result in results['results']:
                    PathScanResult.objects.create(
                        path_scan_job=scan_job,
                        url=url.replace('FUZZ', result['input']['FUZZ']),
                        content_type=result.get('content_type', ''),
                        status=result['status'],
                        length=result['length']
                    )
                scan_job.status = 'C'  # 标记为完成
            else:
                scan_job.status = 'E'
                scan_job.error_message = 'ffuf扫描没有返回结果'
    except subprocess.CalledProcessError as e:
        scan_job.status = 'E'  # 标记为错误
        scan_job.error_message = f'ffuf扫描失败: {e.stderr.decode()}'
    except Exception as e:
        scan_job.status = 'E'  # 标记为错误
        scan_job.error_message = f'处理ffuf扫描结果时发生异常: {str(e)}'
    finally:
        scan_job.end_time = timezone.now()
        scan_job.save()

        # 可选：删除输出文件或保留供后续审查

        if scan_job.status == 'E':
            return {'error': scan_job.error_message}
        return {'message': f'路径扫描完成: {url}'}