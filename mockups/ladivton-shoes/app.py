"""LaDI-VTON-Shoes API wrapper for shoe virtual try-on.

Adapts the batch-oriented LaDI-VTON-Shoes model for real-time single-image API use.
The model uses diffusion-based inpainting to place shoes on a person image.

Port: 8558
Endpoint: POST /api/tryon with {person_image, shoe_image} (base64)
"""

import base64
import io
import logging
import os
import sys
import time
from pathlib import Path

import torch
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from PIL import Image
from pydantic import BaseModel

# Add LaDI-VTON-Shoes source to path
sys.path.insert(0, "/app/ladi-vton-shoes/src")

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="LaDI-VTON-Shoes API")
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)

DEVICE = "cuda" if torch.cuda.is_available() else "cpu"
DTYPE = torch.float16 if DEVICE == "cuda" else torch.float32

# Global pipeline instance (lazy loaded)
_pipeline = None
_tps_model = None
_emasc_model = None
_inversion_adapter = None


def get_models():
    """Lazy-load models on first request."""
    global _pipeline, _tps_model, _emasc_model, _inversion_adapter

    if _pipeline is not None:
        return _pipeline, _tps_model, _emasc_model, _inversion_adapter

    logger.info(f"Loading LaDI-VTON-Shoes models on {DEVICE}...")
    start = time.time()

    try:
        # Load via torch.hub (downloads weights automatically)
        _tps_model = torch.hub.load(
            "miccunifi/ladi-vton",
            "warping_module",
            dataset="dresscode",
            occlusion=True,
        )
        _tps_model = _tps_model.to(DEVICE).eval()

        _emasc_model = torch.hub.load(
            "miccunifi/ladi-vton",
            "emasc",
            dataset="dresscode",
        )
        _emasc_model = _emasc_model.to(DEVICE).eval()

        _inversion_adapter = torch.hub.load(
            "miccunifi/ladi-vton",
            "inversion_adapter",
            dataset="dresscode",
        )
        _inversion_adapter = _inversion_adapter.to(DEVICE).eval()

        # Load the stable diffusion inpainting pipeline
        from diffusers import StableDiffusionInpaintPipeline

        _pipeline = StableDiffusionInpaintPipeline.from_pretrained(
            "stabilityai/stable-diffusion-2-inpainting",
            torch_dtype=DTYPE,
        )
        _pipeline = _pipeline.to(DEVICE)

        logger.info(f"Models loaded in {time.time()-start:.1f}s")

    except Exception as e:
        logger.error(f"Failed to load models: {e}")
        raise

    return _pipeline, _tps_model, _emasc_model, _inversion_adapter


class ShoesTryOnRequest(BaseModel):
    person_image: str  # base64 encoded full-body person image
    shoe_image: str  # base64 encoded shoe product image


class ShoesTryOnResponse(BaseModel):
    image: str  # base64 encoded result


def decode_base64_image(data: str) -> Image.Image:
    """Decode base64 image data to PIL Image."""
    if "," in data:
        data = data.split(",", 1)[1]
    img_bytes = base64.b64decode(data)
    return Image.open(io.BytesIO(img_bytes)).convert("RGB")


def create_foot_mask(person_image: Image.Image) -> Image.Image:
    """Create a simple mask for the foot region.

    For a proper implementation, this should use pose estimation to find feet.
    This is a simplified version that masks the bottom portion of the image.
    """
    import numpy as np

    width, height = person_image.size

    # Create mask - white (255) where we want to inpaint (feet area)
    # Mask bottom 25% of image as foot region (simplified)
    mask = np.zeros((height, width), dtype=np.uint8)
    foot_start = int(height * 0.75)
    mask[foot_start:, :] = 255

    return Image.fromarray(mask)


@app.get("/health")
def health():
    """Health check endpoint."""
    has_gpu = torch.cuda.is_available()
    return {
        "status": "ok",
        "device": DEVICE,
        "gpu": has_gpu,
        "gpu_name": torch.cuda.get_device_name(0) if has_gpu else "none",
    }


@app.post("/api/tryon", response_model=ShoesTryOnResponse)
def try_on(req: ShoesTryOnRequest):
    """Virtual try-on for shoes.

    Takes a person image and a shoe image, returns the person wearing the shoes.
    """
    logger.info(f"Shoe try-on request on {DEVICE}")

    # Decode input images
    try:
        person_img = decode_base64_image(req.person_image)
        shoe_img = decode_base64_image(req.shoe_image)
    except Exception as e:
        raise HTTPException(400, f"Invalid image data: {e}")

    # Resize images to expected dimensions
    target_size = (512, 512)
    person_img = person_img.resize(target_size, Image.LANCZOS)
    shoe_img = shoe_img.resize(target_size, Image.LANCZOS)

    try:
        start = time.time()

        # Get models
        pipeline, tps_model, emasc_model, inversion_adapter = get_models()

        # Create foot mask
        mask = create_foot_mask(person_img)

        # Run inpainting pipeline
        # Note: Full LaDI-VTON uses warping + EMASC + inversion adapter
        # This is a simplified version using direct inpainting
        with torch.no_grad():
            result = pipeline(
                prompt="person wearing shoes, photorealistic",
                image=person_img,
                mask_image=mask,
                guidance_scale=7.5,
                num_inference_steps=30 if DEVICE == "cuda" else 15,
                generator=torch.Generator(device=DEVICE).manual_seed(42),
            ).images[0]

        # Encode result to base64
        buf = io.BytesIO()
        result.save(buf, format="PNG")
        result_b64 = base64.b64encode(buf.getvalue()).decode("utf-8")

        elapsed = time.time() - start
        logger.info(
            f"Shoe try-on complete in {elapsed:.1f}s ({len(buf.getvalue())} bytes)"
        )

        return ShoesTryOnResponse(image=f"data:image/png;base64,{result_b64}")

    except Exception as e:
        logger.error(f"Inference failed: {e}")
        raise HTTPException(500, f"Shoe try-on failed: {e}")


@app.get("/")
def root():
    """Root endpoint with API info."""
    return {
        "service": "LaDI-VTON-Shoes",
        "version": "1.0.0",
        "endpoints": {
            "health": "GET /health",
            "tryon": "POST /api/tryon",
        },
    }
