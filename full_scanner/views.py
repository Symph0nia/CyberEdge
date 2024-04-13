import json

from django.http import JsonResponse
from django.views.decorators.csrf import csrf_exempt
from django.views.decorators.http import require_http_methods

from .tasks import full_scan_subdomains


@csrf_exempt  # 允许跨站请求
@require_http_methods(["POST"])  # 限制只接受POST请求
def full_scan_view(request):
    try:
        # 解析请求体中的JSON
        data = json.loads(request.body.decode('utf-8'))
        targets = data.get('target', '')  # 从请求中获取目标域名字符串
    except json.JSONDecodeError:
        return JsonResponse({'error': '无效的JSON格式'}, status=400)

    # 分割逗号分隔的目标域名字符串，并移除空字符串
    targets_list = [target.strip() for target in targets.split(',') if target.strip()]

    if not targets_list:
        return JsonResponse({'error': '缺少必要的target参数或格式错误'}, status=400)

    task_ids = []
    # 对每个目标域名启动一个子域名扫描任务
    for target in targets_list:
        task = full_scan_subdomains.delay(target)
        task_ids.append(task.id)

    # 返回响应
    return JsonResponse({'message': f'共启动{len(task_ids)}个子域名扫描任务', 'task_ids': task_ids})