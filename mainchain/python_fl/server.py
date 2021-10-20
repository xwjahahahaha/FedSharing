import pickle
import torch
import json
import models
import argparse
import datasets


# 初始化服务器
def init_server(model_name, model_save_path):
    # 初始化模型
    models.init_model(model_name, model_save_path)


# 全局聚合模型
def model_aggregate(conf, global_model, diff_file_path, model_save_path):
    with open(diff_file_path, 'rb') as f:
        weight_accumulator = pickle.load(f)
    # 遍历服务器的全局模型
    for name, data in global_model.state_dict().items():
        # 更新每一层乘上学习率
        update_per_layer = weight_accumulator[name] * conf["lambda"]
        # 累加和
        if data.type() != update_per_layer.type():
            # 因为update_per_layer的type是floatTensor，所以将起转换为模型的LongTensor（有一定的精度损失）
            data.add_(update_per_layer.to(torch.int64))
        else:
            data.add_(update_per_layer)
    torch.save(global_model, model_save_path + conf["model_name"] + ".pth")
    return ">>> model aggregate completed."


# 评估函数
def model_eval(global_model, eval_loader):
    global_model.eval()    # 开启模型评估模式（不修改参数）
    total_loss = 0.0
    correct = 0
    dataset_size = 0
    # 遍历评估数据集合
    for batch_id, batch in enumerate(eval_loader):
        data, target = batch
        # 获取所有的样本总量大小
        dataset_size += data.size()[0]
        # 存储到gpu
        if torch.cuda.is_available():
            data = data.cuda()
            target = target.cuda()
        # 加载到模型中训练
        output = global_model(data)
        # 聚合所有的损失 cross_entropy交叉熵函数计算损失
        total_loss += torch.nn.functional.cross_entropy(
            output,
            target,
            reduction='sum'
        ).item()
        # 获取最大的对数概率的索引值， 即在所有预测结果中选择可能性最大的作为最终的分类结果
        pred = output.data.max(1)[1]
        # 统计预测结果与真实标签target的匹配总个数
        correct += pred.eq(target.data.view_as(pred)).cpu().sum().item()
    acc = 100.0 * (float(correct) / float(dataset_size))    # 计算准确率
    total_1 = total_loss / dataset_size                     # 计算损失值
    print("acc = %f, loss = %f" % (acc, total_1))
    return acc, total_1


if __name__ == '__main__':
    # 设置命令行程序
    parser = argparse.ArgumentParser(description='Federated Learning')
    parser.add_argument('-c', '--conf', dest='conf')
    parser.add_argument('-f', '--func', dest='func', type=int)
    parser.add_argument('-m', '--model-save-path', dest='model_save_path', type=str)
    parser.add_argument('-d', '--diff-save-path', dest='diff_save_path', type=str)
    # 获取所有的参数
    args = parser.parse_args()
    # 读取配置文件
    with open(args.conf, 'r') as f:
        conf = json.load(f)
    if args.func == 1:
        init_server(conf["model_name"], args.model_save_path)
    elif args.func == 2:
        # 加载当前模型
        global_model = torch.load(args.model_save_path + conf["model_name"] + ".pth")
        model_aggregate(conf, global_model, args.diff_save_path, args.model_save_path)
    elif args.func == 3:
        # 加载模型
        global_model = torch.load(args.model_save_path + conf["model_name"] + ".pth")
        # 加载数据集
        _, eval_datasets = datasets.get_dataset("./python_fl/data/", conf["type"])
        # 加载加载器
        eval_loader = torch.utils.data.DataLoader(
            eval_datasets,
            # 设置单个批次大小32
            batch_size=conf["batch_size"],
            # 打乱数据集
            shuffle=True
        )
        model_eval(global_model, eval_loader)
    else:
        print("error func number.")