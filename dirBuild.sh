#!/bin/bash
## 基本上为GPT生成代码

# 默认值为空，表示不指定仓库
selected_repos=""
# 默认输出文件
output_file="repoTimes.csv"
# 默认数据库输出目录
db_output_dir=""
# 日志文件
language="java"  # 默认语言为java
log_file="${language}_log.txt"
# 超时阈值（秒）
timeout_threshold=3600

# 显示用户帮助信息
show_help() {
    echo "Usage: dirBuild.sh [-repo <repo1,repo2,...>] [-o <output_file>] [-l <language>] [-db <db_output_dir>] <root of repos to be analyzed>"
    echo "  -repo   Specify repositories to analyze, separated by commas."
    echo "  -o      Specify the output file for execution times (default: repoTimes.csv)."
    echo "  -l      Specify the programming language (default: java)."
    echo "  -db     Specify the output directory for codeql databases."
    echo "  -h      Display this help message."
    exit 0
}

# 处理命令行选项
while [[ $# -gt 1 ]]; do
    key="$1"

    case $key in
        -repo)
            # 指定要分析的仓库，多个仓库用逗号分隔
            selected_repos="$2"
            shift 2
            ;;
        -o)
            # 指定输出文件
            output_file="$2"
            shift 2
            ;;
        -l)
            # 指定语言
            language="$2"
            log_file="${language}_log.txt"  # 更新日志文件名
            shift 2
            ;;
        -db)
            # 指定数据库输出目录
            db_output_dir="$2"
            shift 2
            ;;
        -h|--help)
            # 显示帮助信息
            show_help
            ;;
        *)
            # 处理其他参数
            echo "Unknown option or argument: $1"
            shift
            ;;
    esac
done

# 检查是否提供了目录路径参数
if [ "$#" -ne 1 ]; then
    echo "请提供目录路径作为参数"
    exit 1
fi

# 获取用户提供的目录路径
root="$1"

# 检查目录是否存在
if [ ! -d "$root" ]; then
    echo "目录不存在: $root"
    exit 1
fi

log_file=$db_output_dir/$log_file

# 获取目录下的所有仓库
repos=""
if [ -z "$selected_repos" ]; then
    # 如果未指定仓库，则获取所有仓库
    repos=$(find "$root" -maxdepth 1 -type d ! -name "." ! -name ".." ! -wholename "$root")
else
    # 如果指定了仓库，则按逗号分隔获取仓库列表
    IFS=',' read -ra repo_array <<< "$selected_repos"
    for repo_name in "${repo_array[@]}"; do
        repo_path="$root/$repo_name"
        # 检查指定的仓库是否存在
        if [ -d "$repo_path" ]; then
            repos+=" $repo_path"
        else
            echo "仓库不存在: $repo_name"
        fi
    done
fi

# 设置数据库输出目录，默认为 $root/../codeql-db
if [ -z "$db_output_dir" ]; then
    db_output_dir="$root/../codeql-db"
fi

# 创建输出文件（覆盖已存在的文件）
echo "repo,执行时间" > "$output_file"
date > "$log_file"

# echo "$repos$"

total_repos=$(echo "$repos" | wc -w)  # 获取总的仓库数量
current_repo=0  # 初始化当前仓库计数

# 初始化失败目录列表
failed_list=()
cleanup_script="$db_output_dir/cleanup_failed_directories.sh"

# 遍历每个仓库并执行codeql命令
for repo in $repos; do
    ((current_repo++))  # 更新当前仓库计数

    # 打印进度条
    progress=$(printf "%.2f" "$(echo "$current_repo / $total_repos * 100" | bc -l)")
    echo -ne "Progress: $progress% ["
    for ((i = 0; i < current_repo; i++)); do
        echo -n "="
    done
    for ((i = current_repo; i < total_repos; i++)); do
        echo -n " "
    done
    echo -ne "]\r"



    start_time=$(date +%s)  # 记录开始时间
    # 使用timeout命令设置超时
    timeout_cmd="timeout --kill-after=60 --signal=TERM $timeout_threshold codeql database create $db_output_dir/$(basename "$repo") -l=$language --overwrite -s=$repo"
    echo "执行命令: $timeout_cmd"
    eval "$timeout_cmd"
    exit_code=$?  # 记录退出码
    end_time=$(date +%s)  # 记录结束时间

    # 计算执行时间并写入输出文件
    if [ $exit_code -eq 0 ]; then
        execution_time=$((end_time - start_time))
        echo "$(basename "$repo"),$(date -u -d @$execution_time +'%H时%M分%S秒')" >> "$output_file"
        echo "$(basename "$repo"),success" >> "$log_file"
    elif [ $exit_code -eq 124 ]; then
        # 如果命令执行超时，则设置执行时间为-1，并写入超时信息
        echo "$(basename "$repo"),-1" >> "$output_file"
        echo "$(basename "$repo"),timeout" >> "$log_file"
        echo "在仓库 $repo 上创建数据库时超时"
        failed_list+=("$(basename "$repo")")
    else
        # 如果命令执行错误，则设置执行时间为-2，并写入失败信息
        echo "$(basename "$repo"),-2" >> "$output_file"
        echo "$(basename "$repo"),fail" >> "$log_file"
        echo "在仓库 $repo 上创建数据库时发生错误"
        failed_list+=("$(basename "$repo")")
    fi
done

# 创建删除失败目录的脚本
{
    echo "#!/bin/bash"
    echo "failed_list=(${failed_list[*]})"
    echo "for dir in \${failed_list[*]}; do"
    echo "    rm -rf \"\$dir\""
    echo "done"
} > "$cleanup_script"
chmod +x "$cleanup_script"

echo "脚本执行完成，执行时间结果已写入: $output_file，日志信息已写入: $log_file，失败仓库删除脚本已写入: $cleanup_script"
