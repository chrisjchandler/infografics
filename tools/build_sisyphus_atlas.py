from __future__ import annotations

import json
import math
from pathlib import Path
from PIL import Image, ImageDraw

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "internal/app/static/assets/generated/sisyphus"
OUT.mkdir(parents=True, exist_ok=True)

# Draw at 64x48 and scale 2x. Every visible pixel is an intentional 2x2 display block.
LW, LH, SCALE = 64, 48, 2
CELL_W, CELL_H = LW * SCALE, LH * SCALE
FRAMES = 8
STATES = ("rising", "steady", "falling")

P = {
    "ink": "#0b1020",
    "outline": "#171d2c",
    "skin_dark": "#7a3f2a",
    "skin": "#c9784f",
    "skin_light": "#f0b37e",
    "hair": "#271b1b",
    "tunic_dark": "#173b5f",
    "tunic": "#246b8e",
    "tunic_light": "#40a7b5",
    "leather": "#6a4b3b",
    "rock_dark": "#3b4658",
    "rock": "#68758a",
    "rock_light": "#aab7c4",
    "rock_glint": "#e8e3d5",
    "dust": "#a26d4f",
    "white": "#f8fafc",
}


def poly(d: ImageDraw.ImageDraw, points, fill, outline=P["outline"]):
    d.polygon([(round(x), round(y)) for x, y in points], fill=fill, outline=outline)


def thick_limb(d: ImageDraw.ImageDraw, points, color, width=4):
    pts = [(round(x), round(y)) for x, y in points]
    d.line(pts, fill=P["outline"], width=width + 2, joint="curve")
    d.line(pts, fill=color, width=width, joint="curve")
    for x, y in (pts[0], pts[-1]):
        r = width // 2
        d.ellipse((x-r, y-r, x+r, y+r), fill=color, outline=P["outline"])


def boulder(d: ImageDraw.ImageDraw, cx: int, cy: int, frame: int, falling: bool):
    r = 12
    d.ellipse((cx-r-1, cy-r-1, cx+r+1, cy+r+1), fill=P["outline"])
    d.ellipse((cx-r, cy-r, cx+r, cy+r), fill=P["rock"])
    # Pixel clusters give the rock volume without blur/gradient.
    poly(d, [(cx-9,cy-7),(cx-4,cy-10),(cx+3,cy-9),(cx+7,cy-5),(cx+2,cy-2),(cx-5,cy-3)], P["rock_light"], None)
    poly(d, [(cx-10,cy+2),(cx-5,cy+8),(cx+2,cy+10),(cx+7,cy+6),(cx+3,cy+2),(cx-3,cy+4)], P["rock_dark"], None)
    d.rectangle((cx-5, cy-9, cx-2, cy-7), fill=P["rock_glint"])
    # Eight authored orientation marks; reverse direction during rollback.
    angle = (-(frame if falling else -frame) * math.pi / 4) + 0.3
    for offset, color in ((0, P["outline"]), (math.pi*0.72, P["rock_dark"])):
        a = angle + offset
        x1 = cx + round(math.cos(a) * 3)
        y1 = cy + round(math.sin(a) * 3)
        x2 = cx + round(math.cos(a) * 9)
        y2 = cy + round(math.sin(a) * 9)
        d.line((x1,y1,x2,y2), fill=color, width=2)
    return cx-r


def draw_frame(state: str, frame: int) -> Image.Image:
    im = Image.new("RGBA", (LW, LH), (0,0,0,0))
    d = ImageDraw.Draw(im)
    phase = frame * math.pi / 4
    stride = round(math.sin(phase) * 3)
    bob = [0,0,1,1,0,0,-1,-1][frame]

    if state == "rising":
        hip = (27, 32+bob)
        shoulder = (31, 22+bob)
        head = (30, 15+bob)
        rock_c = (50, 27+bob)
        rear_foot = (20-stride, 44)
        front_foot = (35+stride, 44)
        arm_upper = [(31,23+bob),(37,22+bob),(40,24+bob)]
        arm_lower = [(30,26+bob),(36,27+bob),(40,29+bob)]
        expression = "strain"
    elif state == "falling":
        hip = (30, 32+bob)
        shoulder = (27, 22+bob)
        head = (25, 15+bob)
        rock_c = (48-frame//3, 28+bob)
        rear_foot = (35+stride, 44)
        front_foot = (23-stride, 44)
        arm_upper = [(27,23+bob),(33,20+bob),(38,24+bob)]
        arm_lower = [(28,27+bob),(34,30+bob),(38,30+bob)]
        expression = "panic"
    else:
        hip = (28, 32+bob)
        shoulder = (30, 23+bob)
        head = (29, 16+bob)
        rock_c = (50, 28+bob)
        rear_foot = (22, 44)
        front_foot = (34, 44)
        arm_upper = [(30,23+bob),(36,22+bob),(40,25+bob)]
        arm_lower = [(30,27+bob),(36,28+bob),(40,30+bob)]
        expression = "brace"

    contact_x = boulder(d, rock_c[0], rock_c[1], frame, state == "falling")

    # Back leg, front leg: broad polygons/limbs rather than disconnected sticks.
    thick_limb(d, [hip, (25-stride,38), rear_foot], P["skin_dark"], 5)
    thick_limb(d, [hip, (31+stride,38), front_foot], P["skin"], 5)
    d.rectangle((rear_foot[0]-4, 43, rear_foot[0]+2, 46), fill=P["leather"])
    d.rectangle((front_foot[0]-3, 43, front_foot[0]+4, 46), fill=P["leather"])

    # Torso/tunic forms a readable forward or backward wedge.
    poly(d, [(shoulder[0]-5,shoulder[1]-2),(shoulder[0]+5,shoulder[1]-1),(hip[0]+5,hip[1]+3),(hip[0]-6,hip[1]+2)], P["tunic"])
    d.line((shoulder[0]-3, shoulder[1]+2, hip[0]+3, hip[1]), fill=P["tunic_light"], width=2)
    d.rectangle((hip[0]-6, hip[1], hip[0]+6, hip[1]+3), fill=P["leather"])

    # Arms are drawn after the torso and terminate exactly on the boulder silhouette.
    upper = [arm_upper[0], arm_upper[1], (contact_x, rock_c[1]-4)]
    lower = [arm_lower[0], arm_lower[1], (contact_x, rock_c[1]+3)]
    thick_limb(d, upper, P["skin_light"], 4)
    thick_limb(d, lower, P["skin"], 4)

    # Neck/head/hair.
    d.rectangle((head[0]-2, head[1]+3, head[0]+2, head[1]+7), fill=P["skin_dark"])
    d.ellipse((head[0]-5,head[1]-5,head[0]+5,head[1]+5), fill=P["outline"])
    d.ellipse((head[0]-4,head[1]-4,head[0]+4,head[1]+4), fill=P["skin"])
    poly(d, [(head[0]-4,head[1]-4),(head[0]+3,head[1]-4),(head[0]+5,head[1]-1),(head[0]-4,head[1])], P["hair"], None)
    d.point((head[0]+3, head[1]), fill=P["ink"])
    if expression == "panic":
        d.line((head[0]+2,head[1]+3,head[0]+4,head[1]+3), fill=P["ink"])
        d.point((head[0]+4,head[1]+4), fill=P["white"])
    elif expression == "strain":
        d.line((head[0]+1,head[1]-1,head[0]+4,head[1]-2), fill=P["ink"])
        d.point((head[0]+4,head[1]+3), fill=P["ink"])
    else:
        d.line((head[0]+2,head[1]+3,head[0]+4,head[1]+2), fill=P["ink"])

    # Sparse state-specific effects.
    if state == "rising" and frame in (1,2,5,6):
        d.rectangle((rear_foot[0]-6,44,rear_foot[0]-4,45), fill=P["dust"])
        d.point((rear_foot[0]-8,43), fill=P["dust"])
    if state == "falling" and frame in (2,3,6,7):
        d.rectangle((rock_c[0]+12,38,rock_c[0]+14,39), fill=P["dust"])

    return im.resize((CELL_W, CELL_H), Image.Resampling.NEAREST)


atlas = Image.new("RGBA", (CELL_W * FRAMES, CELL_H * len(STATES)), (0,0,0,0))
for row, state in enumerate(STATES):
    for frame in range(FRAMES):
        atlas.alpha_composite(draw_frame(state, frame), (frame * CELL_W, row * CELL_H))

atlas_path = OUT / "sisyphus-atlas.png"
atlas.save(atlas_path, optimize=True)

# Poster shows the strongest rising contact pose.
poster = draw_frame("rising", 2)
poster.save(OUT / "sisyphus-poster.png", optimize=True)

metadata = {
    "image": "sisyphus-atlas.png",
    "cell": {"width": CELL_W, "height": CELL_H},
    "frames_per_state": FRAMES,
    "states": {state: {"row": i, "fps": 10 if state != "steady" else 6} for i, state in enumerate(STATES)},
    "anchor": {"x": 56, "y": 92, "meaning": "rear-foot/contact root"},
    "source_inspiration": {
        "file": "../../source/sisyphus/greek-hypaspist.png",
        "license": "CC0-1.0",
        "creator": "Spring Spring",
    },
    "boulder": "Original Infograf artwork; no third-party rock asset used.",
}
(OUT / "sisyphus-atlas.json").write_text(json.dumps(metadata, indent=2) + "\n", encoding="utf-8")
print(atlas_path)
