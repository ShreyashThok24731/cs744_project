import matplotlib.pyplot as plt
import pandas as pd
import os

def plot_single_metric(df, x_col, y_col, title, ylabel, color, filename):
    plt.figure(figsize=(8, 5))
    plt.plot(df[x_col], df[y_col], marker='o', linewidth=2, color=color)
    plt.xlabel('Number of Clients (Load)', fontsize=12)
    plt.ylabel(ylabel, fontsize=12)
    plt.title(title, fontsize=14)
    plt.grid(True, linestyle='--', alpha=0.7)
    plt.tight_layout()
    plt.savefig(filename)
    plt.close()
    print(f"Saved: {filename}")

def plot_utilization(df, workload_name, title_suffix):
    plt.figure(figsize=(8, 5))
    
    plt.plot(df['Clients'], df['Server_CPU(%)'], marker='o', linewidth=2, color='green', label='Server CPU (Core 1)')
    plt.plot(df['Clients'], df['DB_CPU(%)'], marker='^', linewidth=2, color='orange', label='DB CPU (Core 2)')
    plt.plot(df['Clients'], df['DB_IOWait(%)'], marker='x', linestyle='--', linewidth=2, color='black', label='DB Disk Wait')

    plt.xlabel('Number of Clients (Load)', fontsize=12)
    plt.ylabel('Utilization (%)', fontsize=12)
    plt.ylim(-5, 105) 
    plt.title(f"Resource Utilization: {title_suffix}", fontsize=14)
    plt.legend()
    plt.grid(True, linestyle='--', alpha=0.7)
    plt.tight_layout()
    
    filename = f"results/{workload_name}_utilization.png"
    plt.savefig(filename)
    plt.close()
    print(f"Saved: {filename}")

def process_workload(workload_name, title_suffix):
    csv_path = f"results/{workload_name}_results.csv"
    if not os.path.exists(csv_path):
        print(f"File not found: {csv_path}")
        return

    try:
        df = pd.read_csv(csv_path)
    except Exception as e:
        print(f"Error reading CSV: {e}")
        return

    plot_single_metric(
        df, 'Clients', 'Throughput', 
        f"Throughput: {title_suffix}", "Throughput (Req/Sec)", 'tab:blue',
        f"results/{workload_name}_throughput.png"
    )

    plot_single_metric(
        df, 'Clients', 'ResponseTime(ms)', 
        f"Response Time: {title_suffix}", "Response Time (ms)", 'tab:red',
        f"results/{workload_name}_latency.png"
    )

    plot_utilization(df, workload_name, title_suffix)

if __name__ == "__main__":
    process_workload("get_popular", "CPU Bound (Cache Hits)")
    process_workload("put_all", "Disk Bound (Writes)")
