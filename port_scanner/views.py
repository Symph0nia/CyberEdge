import json

from django.http import JsonResponse
from django.views.decorators.csrf import csrf_exempt
from django.views.decorators.http import require_http_methods

from .models import ScanJob, Port
from .tasks import scan_ports  # 确保正确导入异步任务


@csrf_exempt  # 允许跨站请求
@require_http_methods(["POST"])  # 限制只接受POST请求
def scan_ports_view(request):
    try:
        # 解析请求体中的JSON
        data = json.loads(request.body.decode('utf-8'))
        targets = data.get('target', '')
        ports = data.get('ports', '1-65535')  # 如果未指定，设置默认端口范围
    except json.JSONDecodeError:
        return JsonResponse({'error': '无效的JSON格式'}, status=400)

    # 分割逗号分隔的IP地址，并移除空字符串
    targets_list = [ip.strip() for ip in targets.split(',') if ip.strip()]

    if not targets_list:
        return JsonResponse({'error': '缺少必要的target参数或格式错误'}, status=400)

    task_ids = []
    # 对每个IP启动一个任务
    for target in targets_list:
        task = scan_ports.delay(target, ports)
        task_ids.append(task.id)

    # 返回响应
    return JsonResponse({'message': f'共启动{len(task_ids)}个扫描任务', 'task_ids': task_ids})

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
            'ports': list(scan_job.ports.values('id', 'ip_address', 'port_number', 'service_name', 'protocol', 'state')),
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

@csrf_exempt
@require_http_methods(["DELETE"])
def delete_task_view(request, task_id):
    try:
        # 尝试根据提供的task_id找到对应的任务记录
        task = ScanJob.objects.get(task_id=task_id)
        # 删除找到的任务记录
        task.delete()
        return JsonResponse({'message': '任务删除成功'}, status=200)
    except ScanJob.DoesNotExist:
        # 如果没有找到对应的任务记录，则返回错误信息
        return JsonResponse({'error': '任务ID不存在，无法删除'}, status=404)
    except Exception as e:
        # 捕获并处理其他可能的错误
        return JsonResponse({'error': f'删除任务时发生错误: {str(e)}'}, status=500)

@csrf_exempt
@require_http_methods(["DELETE"])
def delete_port_view(request, id):
    try:
        # 尝试根据提供的task_id找到对应的任务记录
        task = Port.objects.get(id=id)
        # 删除找到的任务记录
        task.delete()
        return JsonResponse({'message': '端口删除成功'}, status=200)
    except Port.DoesNotExist:
        # 如果没有找到对应的任务记录，则返回错误信息
        return JsonResponse({'error': '端口ID不存在，无法删除'}, status=404)
    except Exception as e:
        # 捕获并处理其他可能的错误
        return JsonResponse({'error': f'删除端口时发生错误: {str(e)}'}, status=500)