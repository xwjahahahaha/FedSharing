import torch
from torchvision import models
import os


def init_model(model_name, dir):
    if not os.path.exists(dir):
        os.makedirs(dir)
    global_init_model = select_model(model_name)
    torch.save(global_init_model, dir + model_name + ".pth")
    print("success init model.")


def select_model(name="vgg16", pretrained=True):
    if name == "resnet18":
        model = models.resnet18(pretrained=pretrained)
    elif name == "resnet50":
        model = models.resnet50(pretrained=pretrained)
    elif name == "densenet121":
        model = models.densenet121(pretrained=pretrained)
    elif name == "alexnet":
        model = models.alexnet(pretrained=pretrained)
    elif name == "vgg16":
        model = models.vgg16(pretrained=pretrained)
    elif name == "vgg19":
        model = models.vgg19(pretrained=pretrained)
    elif name == "inception_v3":
        model = models.inception_v3(pretrained=pretrained)
    elif name == "googlenet":
        model = models.googlenet(pretrained=pretrained)
    if torch.cuda.is_available():
        return model.cuda()
    else:
        return model



