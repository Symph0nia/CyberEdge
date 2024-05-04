from django.utils import timezone
from celery import shared_task
import subprocess
import json

from common.models import ScanJob
from .models import Subdomain  # 确保正确导入模型
from django.db import transaction

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

    # 构建OneForAll命令
    cmd = f"/OneForAll/oneforall.py --target {target} --fmt json --path {output_file_path} run"

    try:
        # 执行OneForAll命令
        process = subprocess.run(cmd, shell=True, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)

        # 从输出文件读取结果
        with open(output_file_path, 'r') as file:
            results = json.load(file)
            for result in results:
                # 分割和处理包含多个IP的情况
                ip_addresses = result.get('ip', '').split(',')
                # 分割和处理包含多个CNAME的情况
                cnames = result.get('cname', '').split(',')

                for ip in ip_addresses:
                    ip = ip.strip()  # 清除空格
                    if ip:
                        for cname in cnames:
                            cname = cname.strip()  # 清除空格
                            if cname:
                                # 在事务中创建Subdomain对象
                                with transaction.atomic():
                                    Subdomain.objects.create(
                                        scan_job=scan_job,
                                        subdomain=result['subdomain'],
                                        domain=target,
                                        ip_address=ip,
                                        status=result.get('status', ''),
                                        cname=cname,
                                        port=result.get('port', None),
                                        title=result.get('title', ''),
                                        banner=result.get('banner', ''),
                                        addr=result.get('addr', ''),
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

        # 可选：删除输出文件或保留供后续审查

        if scan_job.status == 'E':
            return {'error': scan_job.error_message}
        return {'message': f'子域名扫描完成: {target}'}
