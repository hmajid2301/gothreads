#!/usr/bin/env python3
"""Create NSFW placeholder image for CatVTON safety checker."""
from PIL import Image, ImageDraw, ImageFont

img = Image.new('RGB', (512, 512), color='black')
draw = ImageDraw.Draw(img)
text = 'Content Blocked'

try:
    font = ImageFont.truetype('/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf', 40)
except:
    font = ImageFont.load_default()

bbox = draw.textbbox((0, 0), text, font=font)
x = (512 - (bbox[2] - bbox[0])) // 2
y = (512 - (bbox[3] - bbox[1])) // 2
draw.text((x, y), text, fill='white', font=font)

img.save('/app/catvton/resource/img/NSFW.jpg')
print('Created NSFW placeholder image')
