import torch
import torch.nn as nn
from torchvision import transforms
from PIL import Image
import sys
import os

if len(sys.argv) != 2:
    raise Exception("Please run the file with script and path of image to run inference on")
elif not os.path.exists(sys.argv[1]):
    raise Exception("File does not exist")

# Define the model architecture (ensure it matches the saved model)
class ImprovedPotholeClassifier(nn.Module):
    def __init__(self):
        super(ImprovedPotholeClassifier, self).__init__()
        self.features = nn.Sequential(
            nn.Conv2d(3, 16, kernel_size=3, padding=1),
            nn.ReLU(),
            nn.MaxPool2d(2, 2),
            nn.Conv2d(16, 32, kernel_size=3, padding=1),
            nn.ReLU(),
            nn.MaxPool2d(2, 2),
            nn.Conv2d(32, 64, kernel_size=3, padding=1),
            nn.ReLU(),
            nn.MaxPool2d(2, 2),
            nn.Conv2d(64, 64, kernel_size=3, padding=1),
            nn.ReLU(),
            nn.MaxPool2d(2, 2)
        )
        self.features_output_size = 64 * (224 // 16) * (224 // 16)
        self.classifier = nn.Sequential(
            nn.Linear(self.features_output_size, 512),
            nn.ReLU(),
            nn.Dropout(0.5),
            nn.Linear(512, 2)
        )

    def forward(self, x):
        x = self.features(x)
        x = x.view(-1, self.features_output_size)
        x = self.classifier(x)
        return x

model = ImprovedPotholeClassifier()
model_path = os.path.join(os.path.dirname(__file__), 'best_pothole_model.pth')
model.load_state_dict(torch.load(model_path, map_location=torch.device('cpu'), weights_only=True))

model.eval() 



transform = transforms.Compose([
    transforms.Resize((224, 224)),
    transforms.ToTensor(),
    transforms.Normalize(mean=[0.485, 0.456, 0.406], std=[0.229, 0.224, 0.225])
])


image_path = sys.argv[1]
image = Image.open(image_path).convert('RGB')
input_tensor = transform(image).unsqueeze(0) 

output = model(input_tensor)
_, predicted = torch.max(output, 1)
print(predicted.item())





