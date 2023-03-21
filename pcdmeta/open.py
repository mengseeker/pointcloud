import subprocess

p = subprocess.Popen("./pcdmeta-darwin-arm64", stdin=subprocess.PIPE, stdout=subprocess.PIPE)
for i in range(10):
  p.stdin.writelines([b'{"PCDFile":".dev/compressed.pcd","Labels":[[-1.48,-7.96,-1.62,4.62,1.76,1.98,-7.7908],[29.69,-2.25,-1.555,10.32,4.15,3.26,-7.8208]]}'])
  # p.stdin.writelines([b'{"PCDFile":".dev/notfound.pcd","Labels":[[-1.48,-7.96,-1.62,4.62,1.76,1.98,-7.7908],[29.69,-2.25,-1.555,10.32,4.15,3.26,-7.8208]]}'])
  p.stdin.flush()
  print(p.stdout.readline())

# 关闭标准输出，等待退出子进程
p.stdin.close()
p.wait()
