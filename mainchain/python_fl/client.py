import os
import pickle
import torch
import json
import models
import argparse
import datasets


# 构造函数
def init_client(conf, model_save_path):
    # 客户端初始化模型
    models.init_model(conf["model_name"], model_save_path)


# 模型本地训练函数
def local_train(conf, local_model, model_save_path, diff_save_path, train_loader, epoch, id):
    new_model = models.select_model(conf["model_name"])
    # 更新本地模型，通过部分本地数据集训练得到
    for name, param in local_model.state_dict().items():
        # 客户端首先用服务器端下发的全局模型覆盖本地模型
        new_model.state_dict()[name].copy_(param.clone())
    # 定义最优化函数器用于本地模型训练
    optimizer = torch.optim.SGD(new_model.parameters(), lr=conf['lr'], momentum=conf['momentum'])
    # 本地训练模型
    new_model.train()  # 设置开启模型训练（可以更改参数）
    # 开始训练模型
    print(">>>> Begin training local model.")
    for e in range(conf["local_epochs"]):
        for batch_id, batch in enumerate(train_loader):
            if batch_id % 200 == 0:
                print(">>>> Batch %d completed." % batch_id)
            data, target = batch
            # 加载到gpu
            if torch.cuda.is_available():
                data = data.cuda()
                target = target.cuda()
            # 梯度
            optimizer.zero_grad()
            # 训练预测
            output = new_model(data)
            # 计算损失函数 cross_entropy交叉熵误差
            loss = torch.nn.functional.cross_entropy(output, target)
            # 反向传播
            loss.backward()
            # 更新参数
            optimizer.step()
        print(">>>> Epoch %d done" % (e + 1))
    # 保存模型
    torch.save(new_model, model_save_path + conf["model_name"] + ".pth")
    # 创建差值字典（结构与模型参数同规格），用于记录差值
    diff = dict()
    for name, data in new_model.state_dict().items():
        # 计算训练后与训练前的差值
        diff[name] = (data - local_model.state_dict()[name])
    print(">>>> Client %d local train done" % id)
    # 客户端存储差值
    diff_all_path = diff_save_path + "diff_epoch_" + str(epoch) + ".dict"
    if not os.path.exists(diff_save_path):
        os.makedirs(diff_save_path)
    open(diff_all_path, 'wb').write(pickle.dumps(diff))
    print(">>>> Success save model diff to json file: %s" % diff_all_path)


if __name__ == '__main__':
    # 设置命令行程序
    parser = argparse.ArgumentParser(description='Federated Learning')
    parser.add_argument('-c', '--conf', dest='conf')
    parser.add_argument('-f', '--func', dest='func', type=int)
    parser.add_argument('-m', '--model-save-path', dest='model_save_path', type=str)
    parser.add_argument('-d', '--diff-save-path', dest='diff_save_path', type=str)
    parser.add_argument('-e', '--epoch', dest='epoch', type=int)
    parser.add_argument('-i', '--id', dest='id', type=int)
    # 获取所有的参数
    args = parser.parse_args()
    # 读取配置文件
    with open(args.conf, 'r') as f:
        conf = json.load(f)
    if args.func == 1:
        init_client(conf, args.model_save_path)
    elif args.func == 2:
        # 加载模型
        local_model = torch.load(args.model_save_path + conf["model_name"] + ".pth")
        # 加载数据集
        train_datasets, _ = datasets.get_dataset("./python_fl/data/", conf["type"])
        # 按ID对训练集合的拆分
        all_range = list(range(len(train_datasets)))
        data_len = int(len(train_datasets) / conf['clients_num'])
        indices = all_range[args.id * data_len: (args.id + 1) * data_len]
        # 生成一个数据加载器
        train_loader = torch.utils.data.DataLoader(
            # 制定父集合
            train_datasets,
            # batch_size每个batch加载多少个样本(默认: 1)
            batch_size=conf["batch_size"],
            # 指定子集合
            # sampler定义从数据集中提取样本的策略
            sampler=torch.utils.data.sampler.SubsetRandomSampler(indices)
        )
        local_train(conf, local_model, args.model_save_path, args.diff_save_path, train_loader, args.epoch, args.id)
    else:
        print("error func number.")