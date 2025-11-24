import pandas as pd
import matplotlib.pyplot as plt

df = pd.read_csv("get_popular_results.csv")

fig, ax1 = plt.subplots(figsize=(10, 6))
ax1.set_xlabel("Number of Clients")
ax1.set_ylabel("Throughput (ops/sec)")
ax1.plot(df["Clients"], df["Throughput"], marker='o', color='blue')

ax2 = ax1.twinx()
ax2.set_ylabel("Latency (ms)")
ax2.plot(df["Clients"], df["ResponseTime(ms)"], marker='s', color='red')

plt.title("Throughput & Latency vs Number of Clients")
plt.grid(True)
plt.tight_layout()
plt.savefig("popular_throughput_latency.png")
plt.close()

plt.figure(figsize=(10, 6))
plt.plot(df["Clients"], df["CPU_Utilisation(%)"], marker='o')
plt.xlabel("Number of Clients")
plt.ylabel("CPU Utilization (%)")
plt.title("CPU Utilization vs Number of Clients")
plt.grid(True)
plt.tight_layout()
plt.savefig("popular_cpu.png")
plt.close()

