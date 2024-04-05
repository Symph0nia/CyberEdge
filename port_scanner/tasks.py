from django.utils import timezone
from celery import shared_task
import nmap
from .models import ScanJob, Port  # 确保导入模型

@shared_task(bind=True)  # 添加bind=True以访问self.request.id
def scan_ports(self, target, ports):
    scanner = nmap.PortScanner()
    # 首先创建一个新的ScanJob实例，初始化状态为'R'（Running），并立即保存
    scan_job = ScanJob.objects.create(target=target, status='R', task_id=self.request.id)

    try:
        scanner.scan(target, ports)
        scan_result = scanner[target].get('scan', {})
        if not scan_result:
            scan_job.status = 'E'  # 更新状态为错误
            scan_job.error_message = f'没有扫描结果 {target}。目标可能无法到达或不在线。'
        else:
            # 迭代扫描结果，为每个开放的端口创建Port实例
            for port, port_data in scan_result.get('tcp', {}).items():
                Port.objects.create(
                    scan_job=scan_job,
                    port_number=port,
                    service_name=port_data.get('name', ''),
                    protocol='tcp',
                    state=port_data.get('state', '')
                )
            scan_job.status = 'C'  # 更新状态为完成
    except nmap.PortScannerError as e:
        scan_job.status = 'E'  # 更新状态为错误
        scan_job.error_message = f'Nmap 扫描失败: {str(e)}'
    except KeyError:
        scan_job.status = 'E'  # 更新状态为错误
        scan_job.error_message = f'键错误: 扫描结果不包含 {target} 的预期结构。'
    finally:
        scan_job.end_time = timezone.now()  # 记录结束时间
        scan_job.save()  # 明确保存ScanJob实例的更改

        if scan_job.status == 'E':
            return {'error': scan_job.error_message}
        return {'message': f'扫描完成: {target}'}

