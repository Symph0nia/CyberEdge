from django.utils import timezone
from celery import shared_task
import subprocess
import json
import uuid
from .models import SubdomainScanJob, Subdomain  # 确保正确导入模型

@shared_task(bind=True)
def scan_subdomains(self, target):
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
            for result in results:
                # 创建Subdomain实例
                Subdomain.objects.create(
                    scan_job=scan_job,
                    subdomain=result['subdomain'],
                    ip_address=result.get('ip', ''),
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
    except subprocess.CalledProcessError as e:
        scan_job.status = 'E'  # 标记为错误
        scan_job.error_message = f'子域名扫描失败: {e.stderr.decode()}'
    except Exception as e:
        scan_job.status = 'E'  # 标记为错误
        scan_job.error_message = f'处理子域名扫描结果时发生异常: {str(e)}'
    finally:
        scan_job.end_time = timezone.now()
        scan_job.save()

        # 可选：删除输出文件或保留供后续审查

        if scan_job.status == 'E':
            return {'error': scan_job.error_message}
        return {'message': f'子域名扫描完成: {target}'}
