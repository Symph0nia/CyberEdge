from path_scanner.models import PathScanJob
from subdomain_scanner.models import SubdomainScanJob
from port_scanner.models import PortScanJob
from django.core.exceptions import ObjectDoesNotExist

def get_scan_job_by_task_id(task_id):
    # 假设你的子类名为 PortScanJob, SubdomainScanJob, PathScanJob
    models = [PortScanJob, SubdomainScanJob, PathScanJob]
    for model in models:
        try:
            return model.objects.get(task_id=task_id)
        except ObjectDoesNotExist:
            continue
    return None
