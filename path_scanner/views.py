import json

from django.http import JsonResponse
from django.views.decorators.csrf import csrf_exempt
from django.views.decorators.http import require_http_methods

from .models import PathScanJob, PathScanResult  # 确保正确导入PathScanJob模型
from .tasks import scan_paths  # 确保从你的Celery任务模块导入scan_paths

from django.http import JsonResponse
from django.views.decorators.csrf import csrf_exempt
from django.views.decorators.http import require_http_methods
import json
from .tasks import scan_paths  # 确保正确导入异步任务

@csrf_exempt  # 允许跨站请求
@require_http_methods(["POST"])  # 限制只接受POST请求
def scan_paths_view(request):
    try:
        # 解析请求体中的JSON
        data = json.loads(request.body.decode('utf-8'))
        wordlist = data.get('wordlist', 'default_wordlist.txt')  # 提供默认wordlist文件名
        urls = data.get('url', '')
    except json.JSONDecodeError:
        return JsonResponse({'error': '无效的JSON格式'}, status=400)

    # 分割逗号分隔的URLs，并移除空字符串
    urls_list = [url.strip() for url in urls.split(',') if url.strip()]

    if not urls_list:
        return JsonResponse({'error': '缺少必要的url参数或格式错误'}, status=400)

    task_ids = []
    # 对每个URL启动一个任务
    for url in urls_list:
        task = scan_paths.delay(wordlist, url)
        task_ids.append(task.id)

    # 返回响应
    return JsonResponse({'message': f'共启动{len(task_ids)}个路径扫描任务', 'task_ids': task_ids})

@csrf_exempt
@require_http_methods(["POST"])
def path_task_status_view(request):
    try:
        # 解析请求体中的JSON
        data = json.loads(request.body.decode('utf-8'))
        task_id = data.get('task_id')
    except json.JSONDecodeError:
        return JsonResponse({'error': '无效的JSON格式'}, status=400)

    if not task_id:
        return JsonResponse({'error': '缺少必要的task_id参数'}, status=400)

    # 尝试从数据库获取PathScanJob实例
    try:
        path_scan_job = PathScanJob.objects.get(task_id=task_id)
    except PathScanJob.DoesNotExist:
        return JsonResponse({'error': '任务ID不存在'}, status=404)

    # 构造响应数据
    response_data = {
        'task_id': task_id,
        'task_status': path_scan_job.get_status_display(),
    }

    if path_scan_job.status in ['C', 'E']:  # 如果任务已完成或遇到错误
        response_data['task_result'] = {
            'results': list(path_scan_job.results.values('id', 'url', 'content_type', 'status', 'length')),
            'error_message': path_scan_job.error_message
        }

    return JsonResponse(response_data)

@csrf_exempt
@require_http_methods(["GET"])  # 修改为接受GET请求
def get_all_tasks_view(request):
    # 获取所有ScanJob实例的概要信息
    tasks = PathScanJob.objects.all().values('task_id', 'target', 'status', 'start_time', 'end_time')
    tasks_list = list(tasks)

    # 返回响应
    return JsonResponse({'tasks': tasks_list}, safe=False)  # safe=False允许非字典对象被序列化为JSON

@csrf_exempt
@require_http_methods(["DELETE"])
def delete_task_view(request, task_id):
    try:
        # 尝试根据提供的task_id找到对应的任务记录
        task = PathScanJob.objects.get(task_id=task_id)
        # 删除找到的任务记录
        task.delete()
        return JsonResponse({'message': '任务删除成功'}, status=200)
    except PathScanJob.DoesNotExist:
        # 如果没有找到对应的任务记录，则返回错误信息
        return JsonResponse({'error': '任务ID不存在，无法删除'}, status=404)
    except Exception as e:
        # 捕获并处理其他可能的错误
        return JsonResponse({'error': f'删除任务时发生错误: {str(e)}'}, status=500)

@csrf_exempt
@require_http_methods(["DELETE"])
def delete_path_view(request, id):
    try:
        # 尝试根据提供的task_id找到对应的任务记录
        task = PathScanResult.objects.get(id=id)
        # 删除找到的任务记录
        task.delete()
        return JsonResponse({'message': '端口删除成功'}, status=200)
    except PathScanResult.DoesNotExist:
        # 如果没有找到对应的任务记录，则返回错误信息
        return JsonResponse({'error': '端口ID不存在，无法删除'}, status=404)
    except Exception as e:
        # 捕获并处理其他可能的错误
        return JsonResponse({'error': f'删除端口时发生错误: {str(e)}'}, status=500)