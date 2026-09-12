#!/usr/bin/env python3
"""Parse info_<set> Unity AssetBundle and extract LevelSetInfoSO metadata + screenshots."""

import argparse
import json
import os
import sys

try:
    import UnityPy
    from PIL import Image
except ImportError:
    sys.exit("ERROR: pip install UnityPy Pillow")


def read_field(tree, name, default=""):
    if tree is None:
        return default
    if isinstance(tree, dict):
        return tree.get(name, default)
    return getattr(tree, name, default) if hasattr(tree, name) else default


def read_pptr(tree, name):
    val = read_field(tree, name, None)
    if val is None:
        return None
    if isinstance(val, dict):
        return val.get("m_PathID") or val.get("pathID")
    if hasattr(val, "m_PathID"):
        return val.m_PathID
    return None


def export_texture(env, path_id, out_path):
    if not path_id:
        return False
    try:
        obj = env.objects[int(path_id)]
    except (KeyError, TypeError):
        for o in env.objects:
            if o.path_id == path_id:
                obj = o
                break
        else:
            return False
    data = obj.read()
    img = None
    if hasattr(data, "image"):
        img = data.image
    elif hasattr(data, "m_Texture"):
        tex = data.m_Texture
        if hasattr(tex, "read"):
            tex = tex.read()
        if hasattr(tex, "image"):
            img = tex.image
    if img is None:
        return False
    os.makedirs(os.path.dirname(out_path), exist_ok=True)
    img.save(out_path, "PNG")
    return True


def find_level_set_info(env):
    for obj in env.objects:
        if obj.type.name != "MonoBehaviour":
            continue
        try:
            data = obj.read()
            tree = data.read_typetree() if hasattr(data, "read_typetree") else None
            if tree is None and hasattr(data, "m_Script"):
                tree = data.__dict__
            name = read_field(tree, "levelSetName", "")
            version = read_field(tree, "version", "")
            if name or version:
                return data, tree
        except Exception:
            continue

    for obj in env.objects:
        if obj.type.name == "MonoBehaviour":
            try:
                data = obj.read()
                tree = data.read_typetree()
                if read_field(tree, "levelSetName") or read_field(tree, "version"):
                    return data, tree
            except Exception:
                continue
    return None, None


def find_level_infos(env, level_infos_field):
    results = []
    if not level_infos_field:
        return results
    refs = level_infos_field
    if isinstance(refs, list):
        for ref in refs:
            path_id = None
            if isinstance(ref, dict):
                path_id = ref.get("m_PathID") or ref.get("pathID")
            elif hasattr(ref, "m_PathID"):
                path_id = ref.m_PathID
            if not path_id:
                continue
            for obj in env.objects:
                if obj.path_id != path_id:
                    continue
                try:
                    data = obj.read()
                    tree = data.read_typetree()
                    level_id = read_field(tree, "sceneName", "") or read_field(tree, "levelName", "")
                    level_id = level_id.replace("s_", "") if level_id.startswith("s_") else level_id
                    if not level_id:
                        asset_name = getattr(data, "m_Name", "") or read_field(tree, "m_Name", "")
                        level_id = asset_name.replace("LevelInfo_", "")
                    entry = {
                        "id": level_id,
                        "levelName": read_field(tree, "levelName", ""),
                        "levelNameZh": read_field(tree, "levelNameZH", ""),
                        "sceneName": read_field(tree, "sceneName", ""),
                        "screenshotPathId": read_pptr(tree, "screenshot"),
                    }
                    results.append(entry)
                except Exception:
                    pass
                break
    return results


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--bundle", required=True)
    parser.add_argument("--out-dir", required=True)
    parser.add_argument("--slug", required=True)
    args = parser.parse_args()

    os.makedirs(args.out_dir, exist_ok=True)
    env = UnityPy.load(args.bundle)

    _, tree = find_level_set_info(env)
    if tree is None:
        sys.exit("ERROR: LevelSetInfoSO not found in bundle")

    level_infos = read_field(tree, "levelInfos", [])
    levels = find_level_infos(env, level_infos)

    screenshots_dir = os.path.join(args.out_dir, "screenshots")
    os.makedirs(screenshots_dir, exist_ok=True)

    manifest_levels = []
    for lv in levels:
        shot_rel = ""
        if lv.get("screenshotPathId"):
            shot_path = os.path.join(screenshots_dir, f"{lv['id']}.png")
            if export_texture(env, lv["screenshotPathId"], shot_path):
                shot_rel = f"screenshots/{lv['id']}.png"
        manifest_levels.append({
            "id": lv["id"],
            "levelName": lv.get("levelName", ""),
            "levelNameZh": lv.get("levelNameZh", ""),
            "sceneName": lv.get("sceneName", ""),
            "screenshot": shot_rel,
        })

    manifest = {
        "slug": args.slug,
        "version": read_field(tree, "version", ""),
        "uid": read_field(tree, "uid", ""),
        "levelSetName": read_field(tree, "levelSetName", ""),
        "levelSetNameZH": read_field(tree, "levelSetNameZH", ""),
        "author": read_field(tree, "author", ""),
        "levels": manifest_levels,
    }

    with open(os.path.join(args.out_dir, "manifest.json"), "w", encoding="utf-8") as f:
        json.dump(manifest, f, ensure_ascii=False, indent=2)

    print(json.dumps(manifest, ensure_ascii=False))


if __name__ == "__main__":
    main()
