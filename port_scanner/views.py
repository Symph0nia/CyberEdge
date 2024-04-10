from django.http import JsonResponse
from django.views.decorators.csrf import csrf_exempt
from django.views.decorators.http import require_http_methods
from .models import ScanJob
import json
from .tasks import scan_ports

@csrf_exempt  # 允许跨站请求
@require_http_methods(["POST"])  # 限制只接受POST请求
def scan_ports_view(request):
    try:
        # 解析请求体中的JSON
        data = json.loads(request.body.decode('utf-8'))
        target = data.get('target')
        ports = data.get('ports')
    except json.JSONDecodeError:
        return JsonResponse({'error': '无效的JSON格式'}, status=400)

    if not target or not ports:
        return JsonResponse({'error': '缺少必要的target或ports参数'}, status=400)

    # 异步执行扫描任务
    task = scan_ports.delay(target, ports)

    # 返回响应
    return JsonResponse({'message': '扫描任务已启动', 'task_id': task.id})

@csrf_exempt
@require_http_methods(["POST"])
def task_status_view(request):
    try:
        # 解析请求体中的JSON
        data = json.loads(request.body.decode('utf-8'))
        task_id = data.get('task_id')
    except json.JSONDecodeError:
        return JsonResponse({'error': '无效的JSON格式'}, status=400)

    if not task_id:
        return JsonResponse({'error': '缺少必要的task_id参数'}, status=400)

    # 尝试从数据库获取ScanJob实例
    try:
        scan_job = ScanJob.objects.get(task_id=task_id)
    except ScanJob.DoesNotExist:
        return JsonResponse({'error': '任务ID不存在'}, status=404)

    # 构造响应数据
    response_data = {
        'task_id': task_id,
        'task_status': scan_job.get_status_display(),
    }

    if scan_job.status in ['C', 'E']:  # 如果任务已完成或遇到错误
        response_data['task_result'] = {
            'ports': list(scan_job.ports.values('port_number', 'service_name', 'protocol', 'state')),
            'error_message': scan_job.error_message
        }

    return JsonResponse(response_data)

@csrf_exempt
@require_http_methods(["GET"])  # 修改为接受GET请求
def get_all_tasks_view(request):
    # 获取所有ScanJob实例的概要信息
    tasks = ScanJob.objects.all().values('task_id', 'target', 'status', 'start_time', 'end_time')
    tasks_list = list(tasks)

    # 返回响应
    return JsonResponse({'tasks': tasks_list}, safe=False)  # safe=False允许非字典对象被序列化为JSON