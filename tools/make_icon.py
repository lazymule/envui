#!/usr/bin/env python3
"""Generates the envui icon in the GLARK brand blues.

Run from the repository root (needs Pillow and the DejaVu fonts):

    python3 tools/make_icon.py
    x86_64-w64-mingw32-windres -O coff -i winres/envui.rc -o rsrc_windows_amd64.syso

Writes docs/icon-1024.png, assets/icon.png and winres/envui.ico.
"""
from PIL import Image, ImageDraw, ImageFont, ImageFilter
import sys
S = 1024
SKY, MID, BLUE = (0x9a,0xd3,0xf2), (0x63,0x9c,0xc4), (0x07,0x4e,0x8a)
BAR = (0xe9, 0xf2, 0xf9)
MONO_B = "/usr/share/fonts/truetype/dejavu/DejaVuSansMono-Bold.ttf"

def lerp(a, b, t): return tuple(int(a[i] + (b[i]-a[i])*t) for i in range(3))

def base(margin, radius):
    g = Image.new("RGB", (S, S)); px = g.load()
    for y in range(S):
        for x in range(S):
            t = (x*0.35 + y*0.65) / S
            px[x, y] = lerp(SKY, MID, t/0.45) if t < 0.45 else lerp(MID, BLUE, min(1, (t-0.45)/0.45))
    m = Image.new("L", (S, S), 0)
    ImageDraw.Draw(m).rounded_rectangle([margin, margin, S-margin, S-margin], radius, fill=255)
    out = Image.new("RGBA", (S, S), (0, 0, 0, 0)); out.paste(g, (0, 0), m)
    return out

def icon(simple=False):
    margin = 24 if simple else 48
    img = base(margin, 200 if simple else 210)
    L, T, R, B = (150, 190, 874, 834) if simple else (172, 196, 852, 808)
    if not simple:
        sh = Image.new("RGBA", (S, S), (0, 0, 0, 0))
        ImageDraw.Draw(sh).rounded_rectangle([L+10, T+40, R-10, B+30], 70, fill=(5, 30, 60, 110))
        img.alpha_composite(sh.filter(ImageFilter.GaussianBlur(30)))
    d = ImageDraw.Draw(img)
    rad = 80 if simple else 64
    d.rounded_rectangle([L, T, R, B], rad, fill=(255, 255, 255, 255))
    barH = 150 if simple else 128
    # title bar: blue in the simple icon (reads as a window at 16px), light with ".env" in the full one
    d.rounded_rectangle([L, T, R, T+barH], rad, fill=BLUE if simple else BAR)
    d.rectangle([L, T+barH-rad, R, T+barH], fill=BLUE if simple else BAR)
    if not simple:
        d.text((L+60, T+barH/2+4), ".env", font=ImageFont.truetype(MONO_B, 104), fill=BLUE, anchor="lm")

    top, bottom = T + barH, B
    if simple:
        rows = [(top + (bottom-top)*f) for f in (0.30, 0.70)]
        for i, y in enumerate(rows):
            d.rounded_rectangle([L+80, y-55, L+300, y+55], 55, fill=BLUE)
            d.rounded_rectangle([L+360, y-55, R-80, y+55], 55, fill=MID if i else SKY)
    else:
        rows = [top + (bottom-top)*f for f in (0.24, 0.50, 0.76)]
        spec = [(190, "line"), (140, "dots"), (230, "line")]
        for y, (kw, kind) in zip(rows, spec):
            x = L + 60
            d.rounded_rectangle([x, y-30, x+kw, y+30], 30, fill=BLUE)            # key
            ex = x + kw + 34                                                    # "=" sign
            d.rounded_rectangle([ex, y-20, ex+44, y-6], 7, fill=MID)
            d.rounded_rectangle([ex, y+6, ex+44, y+20], 7, fill=MID)
            vx = ex + 44 + 34
            if kind == "dots":                                                  # masked secret
                for i in range(5):
                    cx = vx + 22 + i*54
                    d.ellipse([cx-19, y-19, cx+19, y+19], fill=MID)
            else:
                d.rounded_rectangle([vx, y-26, R-60, y+26], 26, fill=SKY)
    return img

full, simple = icon(False), icon(True)
full.save("docs/icon-1024.png")
full.resize((512, 512), Image.LANCZOS).save("assets/icon.png", optimize=True)

# multi-resolution .ico: simplified art for small sizes
frames = []
for s in (16, 20, 24, 32, 40, 48, 64, 128, 256):
    src = simple if s <= 32 else full
    frames.append(src.resize((s, s), Image.LANCZOS))
frames[-1].save("winres/envui.ico", format="ICO", sizes=[f.size for f in frames], append_images=frames[:-1])

# preview
sizes = [256, 128, 48, 32, 24, 16]
W = sum(sizes) + 24*len(sizes) + 24
prev = Image.new("RGBA", (W, 2*280), (245, 249, 252, 255))
ImageDraw.Draw(prev).rectangle([0, 280, W, 560], fill=(12, 24, 38, 255))
x = 24
for s in sizes:
    sm = (simple if s <= 32 else full).resize((s, s), Image.LANCZOS)
    prev.alpha_composite(sm, (x, 12)); prev.alpha_composite(sm, (x, 292)); x += s + 24
prev.save("/tmp/envui-icon-preview.png")
