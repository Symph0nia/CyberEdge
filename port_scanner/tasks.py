from django.utils import timezone
from celery import shared_task
import subprocess
import re
from .models import ScanJob, Port  # 确保导入模型

@shared_task(bind=True)
def scan_ports(self, target, ports):
    # 首先创建一个新的ScanJob实例，初始化状态为'R'（Running），并立即保存
    scan_job = ScanJob.objects.create(target=target, status='R', task_id=self.request.id)

    # 创建临时文件名
    temp_file_path = f"/tmp/{scan_job.task_id}.txt"

    try:
        # 构建nmap命令，并将输出重定向到临时文件
        cmd = f"nmap -sS {target} -p {ports} -oN {temp_file_path}"
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
                    Port.objects.create(
                        scan_job=scan_job,
                        port_number=int(port),
                        service_name=service,
                        protocol='tcp',
                        state='open',
                        ip_address=target  # 添加IP地址
                    )
                scan_job.status = 'C'  # 更新状态为完成
    except Exception as e:
        scan_job.status = 'E'  # 更新状态为错误
        scan_job.error_message = f'扫描过程中发生异常: {str(e)}'
    finally:
        scan_job.end_time = timezone.now()  # 记录结束时间
        scan_job.save()  # 明确保存ScanJob实例的更改

        # 可以考虑在此处或在适当的时候删除临时文件，或保留以便审查

        if scan_job.status == 'E':
            return {'error': scan_job.error_message}
        return {'message': f'扫描完成: {target}'}
