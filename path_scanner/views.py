import json
import os

from django.http import JsonResponse
from django.views.decorators.csrf import csrf_exempt
from django.views.decorators.http import require_http_methods

from common.models import ScanJob
from .models import PathScanResult
from .tasks import scan_paths  # 确保正确导入异步任务

@csrf_exempt  # 允许跨站请求
@require_http_methods(["POST"])  # 限制只接受POST请求
def scan_paths_view(request):
    try:
        # 解析请求体中的JSON
        data = json.loads(request.body.decode('utf-8'))
        wordlist = data.get('wordlist', './wordlist/default_wordlist.txt')  # 提供默认wordlist文件名
        urls = data.get('urls', [])  # 直接获取数组格式的URLs
        delay = data.get('delay', 0)
        from_id = data.get('from_id', '')
    except json.JSONDecodeError:
        return JsonResponse({'error': '无效的JSON格式'}, status=400)

    # 确保urls是列表类型
    if not isinstance(urls, list) or not urls:
        return JsonResponse({'error': '缺少必要的urls参数或格式错误'}, status=400)

    # 过滤空字符串URL
    urls_list = [url.strip() for url in urls if url.strip()]

    task_ids = []
    # 对每个URL启动一个任务
    for url in urls_list:
        task = scan_paths.delay(wordlist, url, delay, from_id)
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

    try:
        path_scan_job = ScanJob.objects.get(task_id=task_id)
    except ScanJob.DoesNotExist:
        return JsonResponse({'error': '任务ID不存在'}, status=404)

    # 构造响应数据
    response_data = {
        'task_id': task_id,
        'task_status': path_scan_job.get_status_display(),
    }

    if path_scan_job.status in ['C', 'E']:  # 如果任务已完成或遇到错误
        response_data['task_result'] = {
            'paths': list(path_scan_job.paths.values('id', 'url', 'content_type', 'status', 'length')),
            'error_message': path_scan_job.error_message
        }

    return JsonResponse(response_data)

@csrf_exempt
@require_http_methods(["GET"])
def get_all_tasks_view(request):
    # 获取所有类型为'PATH'的ScanJob实例的概要信息
    tasks = ScanJob.objects.filter(type='PATH')
    tasks_list = []
    for task in tasks:
        tasks_list.append({
            'task_id': task.task_id,
            'target': task.target,
            'status': task.status,
            'result_count': task.result_count,
            'start_time': task.start_time.strftime('%Y年%m月%d日 %H:%M:%S') if task.start_time else None,
            'end_time': task.end_time.strftime('%Y年%m月%d日 %H:%M:%S') if task.end_time else None,
            'from': task.from_job_target,
        })

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

@csrf_exempt
@require_http_methods(["GET"])  # 限制此视图只接受GET请求
def list_wordlists(request):
    wordlist_dir = './wordlist'  # 设置wordlist目录的路径
    try:
        # 获取wordlist目录下的所有文件
        files = []
        for filename in os.listdir(wordlist_dir):
            filepath = os.path.join(wordlist_dir, filename)
            if os.path.isfile(filepath):  # 确保是文件
                files.append({
                    'name': filename,
                    'path': filepath
                })

        # 返回文件列表
        return JsonResponse({'files': files}, status=200)
    except Exception as e:
        # 如果发生错误，返回错误信息
        return JsonResponse({'error': str(e)}, status=500)