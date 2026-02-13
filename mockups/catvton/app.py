"""CatVTON local inference wrapper (CPU).

Runs CatVTON locally on CPU. Used as fallback when HF Space is unavailable.
The Go API calls this service at CATVTON_LOCAL_URL.
Note: CPU inference is slow (~5+ minutes per try-on).
"""

import base64
import io
import logging
import os
import sys
import time

import torch
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from PIL import Image
from pydantic import BaseModel

# Add CatVTON repo to path
sys.path.insert(0, "/app/catvton")

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="CatVTON Local Inference (CPU)")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)

DEVICE = "cuda" if torch.cuda.is_available() else "cpu"
DTYPE = torch.bfloat16 if DEVICE == "cuda" else torch.float32

_pipeline = None


def get_pipeline():
    global _pipeline
    if _pipeline is None:
        from model.pipeline import CatVTONPipeline

        logger.info(f"Loading CatVTON pipeline on {DEVICE} (this may download models on first run)...")
        _pipeline = CatVTONPipeline(
            base_ckpt="runwayml/stable-diffusion-inpainting",
            attn_ckpt="zhengchong/CatVTON",
            attn_ckpt_version="mix",
            weight_dtype=DTYPE,
            use_tf32=(DEVICE == "cuda"),
            device=DEVICE,
        )
        logger.info(f"CatVTON pipeline loaded on {DEVICE}")
    return _pipeline


class TryOnRequest(BaseModel):
    person_image: str
    garment_image: str
    cloth_type: str = "upper"


class TryOnResponse(BaseModel):
    image: str


def decode_base64_image(data: str) -> Image.Image:
    if "," in data:
        data = data.split(",", 1)[1]
    img_bytes = base64.b64decode(data)
    return Image.open(io.BytesIO(img_bytes)).convert("RGB")


@app.get("/health")
def health():
    has_gpu = torch.cuda.is_available()
    return {
        "status": "ok",
        "device": DEVICE,
        "gpu": has_gpu,
        "gpu_name": torch.cuda.get_device_name(0) if has_gpu else "none",
    }


@app.post("/api/tryon", response_model=TryOnResponse)
def try_on(req: TryOnRequest):
    if req.cloth_type not in ("upper", "lower", "overall"):
        raise HTTPException(400, "cloth_type must be upper, lower, or overall")

    logger.info(f"Local try-on request: cloth_type={req.cloth_type}, device={DEVICE}")

    try:
        person_img = decode_base64_image(req.person_image)
        garment_img = decode_base64_image(req.garment_image)
    except Exception as e:
        raise HTTPException(400, f"Invalid image data: {e}")

    target_size = (768, 1024)
    person_img = person_img.resize(target_size, Image.LANCZOS)
    garment_img = garment_img.resize(target_size, Image.LANCZOS)

    mask = Image.new("L", target_size, 255)

    try:
        start = time.time()
        pipeline = get_pipeline()

        # Fewer steps on CPU to reduce wait time
        num_steps = 50 if DEVICE == "cuda" else 20

        result = pipeline(
            image=person_img,
            condition_image=garment_img,
            mask=mask,
            num_inference_steps=num_steps,
            guidance_scale=2.5,
            height=target_size[1],
            width=target_size[0],
            generator=torch.Generator(device=DEVICE).manual_seed(42),
        )

        result_img = result[0] if isinstance(result, (list, tuple)) else result
        if not isinstance(result_img, Image.Image):
            result_img = Image.open(result_img)

        buf = io.BytesIO()
        result_img.save(buf, format="PNG")
        result_b64 = base64.b64encode(buf.getvalue()).decode("utf-8")

        elapsed = time.time() - start
        logger.info(f"Local try-on complete in {elapsed:.1f}s ({len(buf.getvalue())} bytes, {num_steps} steps, {DEVICE})")
        return TryOnResponse(image=f"data:image/png;base64,{result_b64}")

    except Exception as e:
        logger.error(f"Local inference failed: {e}")
        raise HTTPException(500, f"Local try-on failed: {e}")
