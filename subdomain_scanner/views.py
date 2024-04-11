import json

from django.http import JsonResponse
from django.views.decorators.csrf import csrf_exempt
from django.views.decorators.http import require_http_methods

from .models import SubdomainScanJob, Subdomain
from .tasks import scan_subdomains  # 确保从你的Celery任务模块导入scan_subdomains函数


@csrf_exempt  # 允许跨站请求
@require_http_methods(["POST"])  # 限制只接受POST请求
def scan_subdomains_view(request):
    try:
        # 解析请求体中的JSON
        data = json.loads(request.body.decode('utf-8'))
        target = data.get('target')  # 获取要扫描的目标域名
    except json.JSONDecodeError:
        return JsonResponse({'error': '无效的JSON格式'}, status=400)

    if not target:
        return JsonResponse({'error': '缺少必要的target参数'}, status=400)

    # 异步执行OneForAll子域名扫描任务
    task = scan_subdomains.delay(target)

    # 返回响应
    return JsonResponse({'message': '子域名扫描任务已启动', 'task_id': task.id})


@csrf_exempt
@require_http_methods(["POST"])
def subdomain_task_status_view(request):
    try:
        # 解析请求体中的JSON
        data = json.loads(request.body.decode('utf-8'))
        task_id = data.get('task_id')
    except json.JSONDecodeError:
        return JsonResponse({'error': '无效的JSON格式'}, status=400)

    if not task_id:
        return JsonResponse({'error': '缺少必要的task_id参数'}, status=400)

    # 尝试从数据库获取SubdomainScanJob实例
    try:
        subdomain_scan_job = SubdomainScanJob.objects.get(task_id=task_id)
    except SubdomainScanJob.DoesNotExist:
        return JsonResponse({'error': '任务ID不存在'}, status=404)

    # 构造响应数据
    response_data = {
        'task_id': task_id,
        'task_status': subdomain_scan_job.get_status_display(),
    }

    if subdomain_scan_job.status in ['C', 'E']:  # 如果任务已完成或遇到错误
        response_data['task_result'] = {
            'subdomains': list(subdomain_scan_job.subdomains.values('id', 'subdomain', 'ip_address', 'status', 'cname', 'port', 'title',
                    'banner', 'asn', 'org', 'addr', 'isp', 'source')),
            'error_message': subdomain_scan_job.error_message
        }

    return JsonResponse(response_data)

@csrf_exempt
@require_http_methods(["GET"])  # 修改为接受GET请求
def get_all_tasks_view(request):
    # 获取所有ScanJob实例的概要信息
    tasks = SubdomainScanJob.objects.all().values('task_id', 'target', 'status', 'start_time', 'end_time')
    tasks_list = list(tasks)

    # 返回响应
    return JsonResponse({'tasks': tasks_list}, safe=False)  # safe=False允许非字典对象被序列化为JSON

@csrf_exempt
@require_http_methods(["DELETE"])
def delete_task_view(request, task_id):
    try:
        # 尝试根据提供的task_id找到对应的任务记录
        task = SubdomainScanJob.objects.get(task_id=task_id)
        # 删除找到的任务记录
        task.delete()
        return JsonResponse({'message': '任务删除成功'}, status=200)
    except SubdomainScanJob.DoesNotExist:
        # 如果没有找到对应的任务记录，则返回错误信息
        return JsonResponse({'error': '任务ID不存在，无法删除'}, status=404)
    except Exception as e:
        # 捕获并处理其他可能的错误
        return JsonResponse({'error': f'删除任务时发生错误: {str(e)}'}, status=500)

@csrf_exempt
@require_http_methods(["DELETE"])
def delete_domain_view(request, id):
    try:
        # 尝试根据提供的task_id找到对应的任务记录
        task = Subdomain.objects.get(id=id)
        # 删除找到的任务记录
        task.delete()
        return JsonResponse({'message': '子域名删除成功'}, status=200)
    except Subdomain.DoesNotExist:
        # 如果没有找到对应的任务记录，则返回错误信息
        return JsonResponse({'error': '子域名ID不存在，无法删除'}, status=404)
    except Exception as e:
        # 捕获并处理其他可能的错误
        return JsonResponse({'error': f'删除子域名时发生错误: {str(e)}'}, status=500)