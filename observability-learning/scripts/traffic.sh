#!/usr/bin/env bash
set -euo pipefail

rounds="${1:-60}"
if [[ ! "$rounds" =~ ^[1-9][0-9]*$ ]] || (( rounds > 1000 )); then
  echo "用法：bash scripts/traffic.sh [1 到 1000 的轮数]" >&2
  exit 2
fi

for ((i = 1; i <= rounds; i++)); do
  # 500 是故意制造的实验数据，所以这里不用 curl --fail。
  curl --max-time 5 -sS -o /dev/null http://127.0.0.1:18090/hello
  curl --max-time 5 -sS -o /dev/null http://127.0.0.1:18090/slow
  curl --max-time 5 -sS -o /dev/null http://127.0.0.1:18090/error
  sleep 0.2
done
echo "完成 ${rounds} 轮，共 $((rounds * 3)) 个请求。"
