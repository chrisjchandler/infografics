from __future__ import annotations

import json
import math
from pathlib import Path

from PIL import Image, ImageDraw

ROOT = Path(__file__).resolve().parents[1]
ASSET_ROOT = ROOT / "internal/app/static/assets/generated"


def poly(draw: ImageDraw.ImageDraw, points, fill, outline=None):
    draw.polygon([(round(x), round(y)) for x, y in points], fill=fill, outline=outline)


def limb(draw: ImageDraw.ImageDraw, points, color, outline, width):
    pts = [(round(x), round(y)) for x, y in points]
    draw.line(pts, fill=outline, width=width + 2, joint="curve")
    draw.line(pts, fill=color, width=width, joint="curve")
    for x, y in (pts[0], pts[-1]):
        r = max(1, width // 2)
        draw.ellipse((x - r, y - r, x + r, y + r), fill=color, outline=outline)


def save_atlas(name: str, atlas: Image.Image, poster: Image.Image, metadata: dict):
    out = ASSET_ROOT / name
    out.mkdir(parents=True, exist_ok=True)
    atlas.save(out / f"{name}-atlas.png", optimize=True)
    poster.save(out / f"{name}-poster.png", optimize=True)
    (out / f"{name}-atlas.json").write_text(json.dumps(metadata, indent=2) + "\n", encoding="utf-8")


def build_icarus():
    lw, lh, scale = 80, 56, 2
    cell_w, cell_h = lw * scale, lh * scale
    frames = 8
    states = ("rising", "steady", "falling")
    p = {
        "ink": "#0B1020",
        "outline": "#171D2C",
        "skin": "#D7A06A",
        "skin_dark": "#9C6644",
        "hair": "#48271C",
        "tunic_dark": "#173B5F",
        "tunic": "#246B8E",
        "tunic_light": "#40A7B5",
        "wing_dark": "#7B8C9D",
        "wing_mid": "#D7E2EB",
        "wing_light": "#F6F0DA",
        "sun": "#F2B84B",
        "sun_light": "#F8E7A1",
        "alert": "#B42318",
        "cloud": "#E5EEF3",
    }

    def feathered_wing(draw, root_x, root_y, spread, lift, mirror):
        direction = -1 if mirror else 1
        outer = [
            (root_x, root_y),
            (root_x + direction * (16 + spread), root_y - 12 - lift),
            (root_x + direction * (30 + spread), root_y - 8 - lift),
            (root_x + direction * (44 + spread), root_y + 6 - lift // 2),
            (root_x + direction * (54 + spread), root_y + 18),
            (root_x + direction * (28 + spread), root_y + 14),
            (root_x + direction * 10, root_y + 8),
        ]
        poly(draw, outer, p["outline"], None)
        inner = [
            (root_x, root_y + 1),
            (root_x + direction * (14 + spread), root_y - 8 - lift),
            (root_x + direction * (26 + spread), root_y - 5 - lift),
            (root_x + direction * (38 + spread), root_y + 7 - lift // 2),
            (root_x + direction * (44 + spread), root_y + 14),
            (root_x + direction * (24 + spread), root_y + 11),
            (root_x + direction * 8, root_y + 7),
        ]
        poly(draw, inner, p["wing_mid"], None)
        for idx in range(5):
            x1 = root_x + direction * (8 + idx * 7)
            y1 = root_y + 3 + idx
            x2 = root_x + direction * (18 + spread + idx * 6)
            y2 = root_y - 7 - lift + idx * 5
            draw.line((x1, y1, x2, y2), fill=p["wing_dark"], width=2)
        for idx in range(4):
            poly(
                draw,
                [
                    (root_x + direction * (8 + idx * 8), root_y + 4 + idx),
                    (root_x + direction * (14 + idx * 8), root_y + 10 + idx),
                    (root_x + direction * (20 + spread + idx * 7), root_y + 2 + idx * 2),
                    (root_x + direction * (14 + spread + idx * 6), root_y - 4 + idx),
                ],
                p["wing_light"],
                None,
            )

    def frame_image(state: str, frame: int):
        im = Image.new("RGBA", (lw, lh), (0, 0, 0, 0))
        draw = ImageDraw.Draw(im)
        flap = [8, 14, 18, 11, 2, -4, -10, 1][frame]
        bob = [0, -1, -2, -1, 0, 1, 2, 1][frame]
        drift = [0, 1, 2, 1, 0, -1, -2, -1][frame]

        if state == "rising":
            body_x, body_y, tilt = 37 + drift, 26 + bob, -12
            spread, lift = 4, flap
            mood = "euphoric"
        elif state == "falling":
            body_x, body_y, tilt = 41 - drift, 30 + bob, 26
            spread, lift = -8, -10 + flap // 4
            mood = "panic"
        else:
            body_x, body_y, tilt = 39, 28 + bob, -4
            spread, lift = 0, 2 + flap // 3
            mood = "glide"

        feathered_wing(draw, body_x - 4, body_y - 2, spread, lift, True)
        feathered_wing(draw, body_x + 4, body_y - 2, spread, lift, False)

        limb(draw, [(body_x, body_y + 4), (body_x - 6, body_y + 18), (body_x - 12, body_y + 31)], p["skin_dark"], p["outline"], 4)
        limb(draw, [(body_x + 2, body_y + 6), (body_x + 8, body_y + 20), (body_x + 16, body_y + 32)], p["skin"], p["outline"], 4)
        limb(draw, [(body_x - 2, body_y - 1), (body_x - 11, body_y + 12)], p["skin_light"] if "skin_light" in p else p["skin"], p["outline"], 3)
        limb(draw, [(body_x + 2, body_y), (body_x + 12, body_y + 10)], p["skin"], p["outline"], 3)
        poly(
            draw,
            [
                (body_x - 6, body_y - 6),
                (body_x + 7, body_y - 3),
                (body_x + 6, body_y + 14),
                (body_x - 2, body_y + 18),
                (body_x - 8, body_y + 8),
            ],
            p["tunic"],
            p["outline"],
        )
        draw.line((body_x - 3, body_y - 2, body_x + 3, body_y + 10), fill=p["tunic_light"], width=2)
        draw.ellipse((body_x - 5, body_y - 16, body_x + 5, body_y - 6), fill=p["skin"], outline=p["outline"])
        poly(draw, [(body_x - 5, body_y - 16), (body_x + 3, body_y - 16), (body_x + 6, body_y - 12), (body_x - 4, body_y - 10)], p["hair"], None)
        draw.point((body_x + 3, body_y - 13), fill=p["ink"])
        if mood == "panic":
            draw.line((body_x + 2, body_y - 10, body_x + 5, body_y - 8), fill=p["alert"], width=1)
            for idx in range(3):
                draw.rectangle((body_x + 20 + idx * 4, body_y + 6 + idx * 3, body_x + 22 + idx * 4, body_y + 10 + idx * 3), fill=p["sun_light"])
        elif mood == "euphoric":
            draw.line((body_x + 2, body_y - 9, body_x + 5, body_y - 10), fill=p["ink"], width=1)
        else:
            draw.line((body_x + 2, body_y - 9, body_x + 4, body_y - 8), fill=p["ink"], width=1)

        draw.ellipse((56, 6, 71, 21), fill=p["sun"], outline=None)
        draw.ellipse((60, 10, 66, 16), fill=p["sun_light"], outline=None)
        for cloud_x in (8, 20):
            draw.ellipse((cloud_x, 12 + bob, cloud_x + 10, 18 + bob), fill=p["cloud"], outline=None)
            draw.ellipse((cloud_x + 6, 9 + bob, cloud_x + 17, 18 + bob), fill=p["cloud"], outline=None)

        return im.resize((cell_w, cell_h), Image.Resampling.NEAREST)

    atlas = Image.new("RGBA", (cell_w * frames, cell_h * len(states)), (0, 0, 0, 0))
    for row, state in enumerate(states):
        for frame in range(frames):
            atlas.alpha_composite(frame_image(state, frame), (frame * cell_w, row * cell_h))

    metadata = {
        "image": "icarus-atlas.png",
        "cell": {"width": cell_w, "height": cell_h},
        "frames_per_state": frames,
        "states": {
            "rising": {"row": 0, "fps": 10},
            "steady": {"row": 1, "fps": 6},
            "falling": {"row": 2, "fps": 10},
        },
        "anchor": {"x": 78, "y": 58, "meaning": "pelvis/chest root"},
        "provenance": "Original Infograf artwork guided by public-domain Icarus pose studies listed in docs/ASSET_LICENSES.md.",
    }
    save_atlas("icarus", atlas, frame_image("rising", 2), metadata)


def build_cerberus():
    lw, lh, scale = 96, 72, 2
    cell_w, cell_h = lw * scale, lh * scale
    frames = 8
    states = ("rising", "steady", "falling")
    p = {
        "ink": "#0B1020",
        "outline": "#171D2C",
        "fur_dark": "#0E3040",
        "fur": "#15535A",
        "fur_light": "#238073",
        "tooth": "#F4F0E8",
        "eye": "#D0E85B",
        "tongue": "#F06A5B",
        "alert": "#B93B4B",
        "shadow": "#0A1E2B",
    }

    def head(draw, cx, cy, open_jaw, horn, glance):
        poly(
            draw,
            [(cx - 10, cy - 4), (cx + 2, cy - 12), (cx + 16, cy - 8), (cx + 18, cy + 4), (cx + 8, cy + 12), (cx - 8, cy + 9)],
            p["outline"],
            None,
        )
        poly(
            draw,
            [(cx - 8, cy - 3), (cx + 2, cy - 10), (cx + 13, cy - 7), (cx + 15, cy + 3), (cx + 6, cy + 10), (cx - 6, cy + 7)],
            p["fur"],
            None,
        )
        if horn:
            poly(draw, [(cx - 2, cy - 12), (cx + 1, cy - 22), (cx + 5, cy - 12)], p["alert"], None)
        draw.rectangle((cx + 3, cy - 4, cx + 7, cy - 2), fill=p["eye"])
        if open_jaw:
            poly(draw, [(cx + 1, cy + 5), (cx + 18, cy + 6), (cx + 14, cy + 14), (cx + 2, cy + 11)], p["shadow"], None)
            draw.rectangle((cx + 5, cy + 11, cx + 12, cy + 13), fill=p["tongue"])
            for tooth_x in (8, 12, 16):
                draw.rectangle((cx + tooth_x, cy + 6, cx + tooth_x + 1, cy + 9), fill=p["tooth"])
        else:
            draw.line((cx + 1, cy + 8, cx + 14, cy + 10), fill=p["shadow"], width=2)
        draw.line((cx + 2, cy - 8 + glance, cx + 8, cy - 9), fill=p["shadow"], width=1)

    def frame_image(state: str, frame: int):
        im = Image.new("RGBA", (lw, lh), (0, 0, 0, 0))
        draw = ImageDraw.Draw(im)
        bob = [0, 1, 2, 1, 0, -1, -2, -1][frame]
        stride = [0, 2, 4, 2, 0, -2, -4, -2][frame]
        if state == "rising":
            lift = -3
            jaw = frame in (1, 2, 5, 6)
        elif state == "falling":
            lift = 4
            jaw = frame in (0, 3, 4, 7)
        else:
            lift = 0
            jaw = frame in (2, 6)

        poly(
            draw,
            [(18, 39 + bob), (30, 26 + lift), (50, 20 + lift), (66, 22 + lift), (80, 32 + bob), (76, 46 + bob), (58, 53 + bob), (30, 51 + bob)],
            p["outline"],
            None,
        )
        poly(
            draw,
            [(20, 40 + bob), (31, 28 + lift), (50, 22 + lift), (64, 24 + lift), (77, 33 + bob), (73, 44 + bob), (57, 50 + bob), (31, 49 + bob)],
            p["fur"],
            None,
        )
        poly(draw, [(34, 27 + lift), (49, 23 + lift), (61, 25 + lift), (65, 35 + bob), (52, 41 + bob), (37, 39 + bob)], p["fur_light"], None)

        limb(draw, [(28, 40 + bob), (24 + stride, 54), (20 + stride, 68)], p["fur_dark"], p["outline"], 6)
        limb(draw, [(42, 42 + bob), (40 - stride // 2, 56), (38 - stride // 2, 69)], p["fur"], p["outline"], 6)
        limb(draw, [(59, 42 + bob), (63 + stride // 2, 57), (69 + stride // 2, 69)], p["fur"], p["outline"], 6)
        limb(draw, [(72, 38 + bob), (76 - stride, 54), (82 - stride, 67)], p["fur_dark"], p["outline"], 6)

        necks = [
            [(34, 24 + lift), (28, 12 + bob), (26, 4 + bob)],
            [(52, 21 + lift), (52, 9 + bob), (52, 1 + bob)],
            [(67, 24 + lift), (74, 11 + bob), (79, 5 + bob)],
        ]
        for idx, neck in enumerate(necks):
            limb(draw, neck, p["fur"], p["outline"], 6)
            head(draw, neck[-1][0], neck[-1][1], jaw and idx != 1, idx == 1, (-1 + idx))

        draw.line((78, 35 + bob, 90, 26 + bob), fill=p["outline"], width=4)
        draw.line((78, 35 + bob, 89, 27 + bob), fill=p["fur_light"], width=2)
        return im.resize((cell_w, cell_h), Image.Resampling.NEAREST)

    atlas = Image.new("RGBA", (cell_w * frames, cell_h * len(states)), (0, 0, 0, 0))
    for row, state in enumerate(states):
        for frame in range(frames):
            atlas.alpha_composite(frame_image(state, frame), (frame * cell_w, row * cell_h))

    metadata = {
        "image": "cerberus-atlas.png",
        "cell": {"width": cell_w, "height": cell_h},
        "frames_per_state": frames,
        "states": {
            "rising": {"row": 0, "fps": 9},
            "steady": {"row": 1, "fps": 6},
            "falling": {"row": 2, "fps": 9},
        },
        "anchor": {"x": 90, "y": 98, "meaning": "body root"},
        "provenance": "Original Infograf artwork for the Cerberus option.",
    }
    save_atlas("cerberus", atlas, frame_image("steady", 2), metadata)


def main():
    build_icarus()
    build_cerberus()


if __name__ == "__main__":
    main()
