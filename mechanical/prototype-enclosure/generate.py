"""DepthPilot panel enclosure concept v1. All geometry in mm; run from any cwd."""
from pathlib import Path
import json
import manifold3d as m
import numpy as np
import trimesh

OUT = Path(__file__).parent
STL = OUT / 'stl'
STL.mkdir(exist_ok=True)
W, D, H, T = 440, 460, 228, 4
parts, placed, report = {}, [], {}

def box(x, y, z, p=(0, 0, 0)):
    return m.Manifold.cube((x, y, z)).translate(p)

def hole(x, y, r=2.25, depth=8):
    return m.Manifold.cylinder(depth, r, r, 48).translate((x, y, -1))

def stations(n):
    return (48, n / 2, n - 48)

def panel(a, b):
    s = box(a, b, T)
    for edge in (14, a - 14):
        for q in stations(b):
            for delta in (-8, 8):
                s -= hole(edge, q + delta)
    for edge in (14, b - 14):
        for q in stations(a):
            for delta in (-8, 8):
                s -= hole(q + delta, edge)
    return s

def vents(s, x0, y0, nx, ny):
    for i in range(nx):
        for j in range(ny):
            s -= box(5, 18, 6, (x0 + 10*i, y0 + 24*j, -1))
    return s

def mesh(s):
    a = s.to_mesh()
    return trimesh.Trimesh(np.asarray(a.vert_properties)[:, :3],
                           np.asarray(a.tri_verts), process=True)

def save(name, s, qty):
    assert s.status() == m.Error.NoError and len(s.decompose()) == 1, name
    tri = mesh(s)
    assert tri.is_volume and tri.is_watertight, name
    assert max(tri.extents) <= 240.01 and tri.bounds[0, 2] >= -0.001, name
    path = STL / (name + '.stl')
    tri.export(path)
    check = trimesh.load_mesh(path)
    assert check.is_volume and check.is_watertight, name
    parts[name] = s
    report[name] = dict(quantity=qty, bounds_mm=tri.extents.round(3).tolist(),
                        volume_cm3=round(tri.volume/1000, 2), triangles=len(tri.faces))

def put(name, rotation, offset):
    placed.append((name, parts[name].rotate(rotation).translate(offset)))

save('floor-tile', panel(220, 230), 4)
save('roof-tile', vents(panel(220, 230), 45, 45, 14, 6), 4)
front = panel(220, 220) - box(64,64,8,(118,64,-1))
save('front-left', front, 1)
save('front-right', panel(220, 220) - box(64,64,8,(38,64,-1)), 1)
rear = vents(panel(220, 220), 40, 35, 15, 6) + box(192,25,4,(14,85,0))
for x in (20,110,180):
    rear -= hole(x,96)
save('rear-tile', rear, 2)
save('side-tile', vents(panel(226, 220), 40, 35, 15, 6), 4)
# 80 mm fan: 71.5 mm mounting pitch; grille remains integral.
fan = panel(226, 220)
for x in (113-35.75, 113+35.75):
    for y in (110-35.75, 110+35.75):
        fan -= hole(x, y)
for x in range(80, 146, 10):
    fan -= box(5, 58, 6, (x, 81, -1))
save('side-fan-tile', fan, 0)  # optional substitution for a side tile
# Splice: align its centre with any seam station; plain M4 nuts, no inserts.
splice = box(48, 32, 4)
for x in (10, 38):
    for y in (8, 24):
        splice -= hole(x, y)
save('seam-splice', splice, 24)
# Butt-joint angle: floor/overlapping face uses 10 mm leg; inset face 14 mm.
angle = box(24, 32, 4) + box(4, 32, 24)
for lo, hi in ((4, 12), (20, 28)):
    # Rounded slots allow the 4 mm station shift at inset wall edges.
    cut = hole(10, lo) + hole(10, hi) + box(4.5, hi-lo, 8, (7.75, lo, -1))
    angle -= cut
    cut = hole(14, lo) + hole(14, hi) + box(4.5, hi-lo, 8, (11.75, lo, -1))
    angle -= cut.rotate((0, -90, 0)).translate((4, 0, 0))
# Print extrusion axis vertical: no long unsupported angle ledge.
angle = angle.rotate((90, 0, 0)).translate((0, 24, 0))
save('corner-angle', angle, 40)
save('rail-spacer', m.Manifold.cylinder(8,8,8,48) - hole(0,0,depth=10), 6)
# Flat drill-to-fit pads attach through selected floor holes after drilling.
save('equipment-pad', box(160, 120, 6), 2)
save('fit-coupon', panel(60, 60) - hole(30, 30, 7.15), 1)

for ix in range(2):
    for iy in range(2):
        put('floor-tile', (0,0,0), (220*ix,230*iy,0))
        put('roof-tile', (0,0,0), (220*ix,230*iy,224))
put('front-left', (90,0,0), (0,4,4))
put('front-right', (90,0,0), (220,4,4))
for ix in range(2):
    put('rear-tile', (90,0,0), (220*ix,460,4))
for iy in range(2):
    put('side-tile', (90,0,90), (0,4+226*iy,4))
    put('side-tile', (90,0,90), (436,4+226*iy,4))
for i,(name,s) in enumerate(placed):
    for other,t in placed[:i]:
        assert (s ^ t).volume() < .001, ('Panel overlap',name,other)

# Layout envelopes, NOT vendor CAD. Coordinates are assembly X/Y/Z.
components = [
    ('motor-left', (121.5, 0, 71.5), (57,195,57), '#edaa43'),
    ('motor-right', (261.5, 0, 71.5), (57,195,57), '#edaa43'),
    ('driver-left-reservation', (25,45,35), (45,155,120), '#ac85d1'),
    ('driver-right-reservation', (365,45,35), (45,155,120), '#ac85d1'),
    ('Jetson-with-connectors-reservation', (75,210,20), (145,130,85), '#52b6c7'),
    ('NDR-240-48', (345,326.5,40), (63,113.5,125.2), '#63b685'),
    ('12V-converter-reservation', (30,335,40), (40,105,125), '#85c59b'),
    ('24V-converter-reservation', (85,375,40), (25,65,100), '#85c59b'),
    ('brake-relays-reservation', (125,360,40), (20,80,100), '#85c59b'),
]
scene = trimesh.Scene()
for index, (name, s) in enumerate(placed):
    tri = mesh(s)
    tri.visual.vertex_colors = [165,180,200,85]
    scene.add_geometry(tri, node_name=f'{name}-{index}')
for name,p,size,color in components:
    s = box(*size,p)
    # Confirm reservations clear every printed panel and one another.
    for panel_name, panel_s in placed:
        assert (s ^ panel_s).volume() < .001, (name,panel_name)
    tri=mesh(s)
    tri.visual.vertex_colors=trimesh.visual.color.hex_to_rgba(color)
    scene.add_geometry(tri,node_name=name)
for i, (name,p,size,_) in enumerate(components):
    for other,q,extent,_ in components[:i]:
        assert (box(*size,p) ^ box(*extent,q)).volume() < .001, (name,other)
for x in (150,290):
    # External drum envelope, rear face 3 mm ahead of flush gearbox face.
    drum = m.Manifold.cylinder(50,55,55,96).rotate((90,0,0)).translate((x,-3,100))
    tri = mesh(drum)
    tri.visual.vertex_colors = [235,165,65,180]
    scene.add_geometry(tri,node_name=f'external-drum-envelope-{x}')
scene.apply_scale(.001)  # glTF convention: metres; source and STLs remain mm.
scene.export(OUT/'layout.glb')
# Assembly shell only; separate shells intentional, not intended as a print.
trimesh.util.concatenate([mesh(s) for _,s in placed]).export(OUT/'assembly-reference.stl')
(OUT/'validation.json').write_text(json.dumps(report,indent=2)+'\n')
(OUT/'component-envelopes.json').write_text(json.dumps([
    dict(name=n,origin_mm=p,size_mm=s,color=c) for n,p,s,c in components],indent=2)+'\n')

# Portable vector drawing with top and front orthographic views.
svg=['<svg xmlns="http://www.w3.org/2000/svg" width="1100" height="720" viewBox="0 0 1100 720">',
     '<style>text{font-family:Arial;font-size:13px} .title{font-size:22px;font-weight:bold}</style>',
     '<rect width="1100" height="720" fill="#fff"/>',
     '<text x="30" y="32" class="title">DepthPilot prototype enclosure — layout v1 (mm)</text>',
     '<text x="30" y="58">440 W × 460 D × 228 H; two external Ø110 drums; print panels separately</text>',
     '<g transform="translate(30 95)"><rect width="440" height="460" fill="#eef1f5" stroke="#222"/>']
for index,(n,(x,y,z),(a,b,c),color) in enumerate(components,1):
    svg += [f'<rect x="{x}" y="{y}" width="{a}" height="{b}" fill="{color}" stroke="#333"/>',
            f'<text x="{x+3}" y="{y+16}">{index}</text>']
svg += ['<text x="155" y="-10">FRONT / DRUM SIDE</text>',
        '<path d="M220 0V460 M0 230H440" stroke="#555" stroke-dasharray="5 5" fill="none"/>',
        '</g><g transform="translate(580 95)">',
        '<rect width="440" height="228" fill="#eef1f5" stroke="#222"/>']
for x in (150,290):
    svg += [f'<circle cx="{x}" cy="128" r="55" fill="#edaa43" fill-opacity=".5" stroke="#222"/>',
            f'<rect x="{x-32}" y="96" width="64" height="64" fill="none" stroke="#222"/>']
svg += ['<text x="100" y="260">Shaft centres: X=150 / 290, Z=100</text>',
        '<text x="100" y="282">140 mm spacing; 30 mm drum-to-drum gap</text></g>',
        '<text x="580" y="410">Component envelope legend:</text>']
for index,(name,_,__,color) in enumerate(components,1):
    svg += [f'<rect x="580" y="{420+index*19}" width="12" height="12" fill="{color}"/>',
            f'<text x="600" y="{431+index*19}">{index}. {name}</text>']
svg += ['<text x="30" y="635">Dashed lines: floor seams. Covers omitted in top view. Component blocks are reservations, not mounting templates.</text>',
        '<text x="30" y="657">Motor support uses metal brackets/base. See README for mounting, print quantities and unresolved physical fit checks.</text></svg>']
(OUT/'layout.svg').write_text('\n'.join(svg))
print(json.dumps(report,indent=2))
