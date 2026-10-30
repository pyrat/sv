"""Outdoor enclosure development kit: independent gasketed cells, mm.

Printed covers close interfaces for leak trials. Working shaft cartridges and
thermal interfaces are separate engineering items, not simulated by these caps.
"""
from pathlib import Path
import json
import zipfile
import manifold3d as m
import numpy as np
import trimesh

OUT = Path(__file__).parent
STL = OUT/'stl'
STL.mkdir(exist_ok=True)
report = {}
parts = {}
INSERT_D = 5.8  # provisional M4 heat-set pocket; match purchased insert
GASKET_H = 2.4
GROOVE_H = 2.0

def box(size, p=(0,0,0)):
    return m.Manifold.cube(size).translate(p)

def ring(outer, inner, height, p=(0,0,0)):
    d = (outer-inner)/2
    return (box((outer,outer,height))-box((inner,inner,height+2),(d,d,-1))).translate(p)

def cyl(r,h,p=(0,0,0)):
    return m.Manifold.cylinder(h,r,r,64).translate(p)

def yhole(r,h,x,y,z):
    return cyl(r,h).rotate((-90,0,0)).translate((x,y,z))

def mesh(s):
    a = s.to_mesh()
    return trimesh.Trimesh(np.asarray(a.vert_properties)[:,:3],np.asarray(a.tri_verts),process=True)

def save(name,s,qty,material):
    assert s.status()==m.Error.NoError and len(s.decompose())==1,name
    tri=mesh(s)
    assert tri.is_volume and tri.is_watertight,name
    assert max(tri.extents)<=240.001 and tri.bounds[0,2]>=-.001,name
    path=STL/f'{name}.stl'
    tri.export(path)
    check=trimesh.load_mesh(path)
    assert check.is_volume and check.is_watertight,name
    parts[name]=s
    report[name]=dict(quantity=qty,material=material,bounds_mm=tri.extents.round(3).tolist(),
                     solid_volume_cm3=round(tri.volume/1000,2),triangles=len(tri.faces))

# Eight screws are outside the gasket; all body pockets terminate blind.
lid_bolts=[(7,7),(115,7),(223,7),(223,115),(223,223),(115,223),(7,223),(7,115)]
port_bolts=[(53,43),(115,43),(177,43),(177,105),(177,167),(115,167),(53,167),(53,105)]
body=box((230,230,220))-box((218,218,214),(6,6,8))
body += ring(230,186,12,(0,0,208))
# Continuous 4 mm wide groove; gasket is 2.8 mm wide, leaving displacement room.
body -= ring(202,194,GROOVE_H+.1,(14,14,220-GROOVE_H))
for x,y in lid_bolts:
    body -= cyl(INSERT_D/2,6.1,(x,y,214))
# Reinforced service interface: replace cap with machined shaft or thermal hardware.
body += box((140,12,140),(45,0,35))
body -= box((100,14,100),(65,-1,55))
port_groove=ring(116,108,2.1).rotate((-90,0,0)).translate((57,-.1,163))
body -= port_groove
for x,z in port_bolts:
    body -= yhole(INSERT_D/2,6,x,0,z)
# Internal equipment pads: blind pilot bores; no enclosure-wall penetration.
for x in (35,115,195):
    for y in (40,115,190):
        body += cyl(7,6,(x,y,8))
        body -= cyl(1.7,5.1,(x,y,9))
# Underside attachment pockets stop 3 mm short of the inside floor.
for x in (30,200):
    for y in (30,200):
        body -= cyl(INSERT_D/2,5,(x,y,0))
save('cell-body',body,4,'ASA preferred; PETG trial')

# Lid printed upside down: top on bed, skirt upwards. Body clearance 0.5/side.
lid=box((238,238,6))+ring(238,231,10,(0,0,6))
for x,y in lid_bolts:
    lid -= cyl(2.25,7,(x+4,y+4,-.5))
save('rain-skirt-lid',lid,4,'ASA preferred; PETG trial')
save('lid-gasket-tpu',ring(200.8,195.2,GASKET_H),4,'TPU; calibrate first')

cap=box((136,136,6))
for x,z in port_bolts:
    cap -= cyl(2.25,8,(x-47,z-37,-1))
save('interface-blank-cap',cap,4,'Rigid; leak-test blank only')
save('interface-gasket-tpu',ring(114.8,109.2,GASKET_H),4,'TPU; calibrate first')

# A small sealed test chamber uses exactly the production groove cross-section.
coupon=box((70,70,12))-box((30,30,5),(20,20,8))
coupon -= ring(48,40,2.1,(11,11,10))
for x in (6,64):
    for y in (6,64):
        coupon -= cyl(INSERT_D/2,6.1,(x,y,6))
save('seal-test-base',coupon,1,'Rigid')
test_lid=box((70,70,6))
for x in (6,64):
    for y in (6,64):
        test_lid -= cyl(2.25,8,(x,y,-1))
save('seal-test-lid',test_lid,1,'Rigid')
for thickness in (2.2,2.4,2.6):
    save(f'seal-test-gasket-{thickness:.1f}mm',ring(46.8,41.2,thickness),1,'TPU')

# Demonstrate actual closed stack, with a geometry surrogate for compressed TPU.
assembled_lid=lid.rotate((180,0,0)).translate((-4,234,226))
assert (body ^ assembled_lid).volume()<.001,'Lid hits body'
lid_seal=ring(200.8,195.2,2,(14.6,14.6,218))
assert (body ^ lid_seal).volume()<.001,'Lid seal does not fit groove'
cap_installed=cap.rotate((90,0,0)).translate((47,0,37))
assert (body ^ cap_installed).volume()<.001,'Cap hits body'
port_seal=ring(114.8,109.2,2).rotate((-90,0,0)).translate((57.6,0,162.4))
assert (body ^ port_seal).volume()<.001,'Port seal does not fit groove'
# The full pressure displacement must fit within the rectangular gland area.
assert 2.8*GASKET_H < 4*GROOVE_H

scene=trimesh.Scene()
def add(s,name,color):
    t=mesh(s)
    t.visual.vertex_colors=color
    scene.add_geometry(t,node_name=name)

# 30 mm service gap between cells; covers overhang 4 mm per side.
for index,(x,y) in enumerate(((0,0),(260,0),(0,260),(260,260)),1):
    for s,name,color in ((body,'body',[150,165,185,150]),
                         (assembled_lid,'lid',[180,195,210,120]),
                         (lid_seal,'compressed-lid-gasket',[25,180,130,255]),
                         (cap_installed,'interface-cap',[225,155,60,255]),
                         (port_seal,'compressed-port-gasket',[25,180,130,255])):
        # Rear interface caps face rear, leaving front-to-rear service gap clear.
        if index>2:
            s=s.rotate((0,0,180)).translate((230,230,0))
        add(s.translate((x,y,0)),f'cell-{index}-{name}',color)
scene.apply_scale(.001)
scene.export(OUT/'assembly.glb')
(OUT/'validation.json').write_text(json.dumps(dict(
    parts=report,checks=['single connected solid per STL','watertight export reload',
        'positive volume','print extents <=240 mm','lid/body clearance',
        'lid and port gasket/groove fit','cap/body clearance','gland displacement area'],
    physical_water_test='NOT PERFORMED',thermal_test='NOT PERFORMED',ip_rating='NONE'),indent=2)+'\n')

svg=['<svg xmlns="http://www.w3.org/2000/svg" width="1100" height="780" viewBox="0 0 1100 780">',
     '<style>text{font-family:Arial;font-size:15px} .title{font-size:24px;font-weight:bold}</style>',
     '<rect width="1100" height="780" fill="white"/>',
     '<text x="30" y="35" class="title">DepthPilot outdoor enclosure development — four sealed cells</text>',
     '<text x="30" y="65">Bodies 230 × 230 × 220 mm; lids 238 × 238 mm; no open ventilation slots</text>']
names=['Motor / driver A','Motor / driver B','Jetson + 12 V converter','48 V PSU + 24 V / relays']
for i,(x,y) in enumerate(((30,110),(290,110),(30,370),(290,370))):
    svg += [f'<rect x="{x}" y="{y}" width="230" height="230" fill="#edf1f5" stroke="#333"/>',
            f'<rect x="{x+14}" y="{y+14}" width="202" height="202" fill="none" stroke="#19a67b" stroke-width="4"/>',
            f'<text x="{x+18}" y="{y+110}">{names[i]}</text>',
            f'<text x="{x+18}" y="{y+135}">Independent sealed volume</text>']
svg += ['<text x="30" y="100">FRONT: external drums; shaft cartridges still to be engineered</text>',
        '<text x="570" y="125" class="title">Static seal section (not to scale)</text>',
        '<rect x="600" y="160" width="400" height="35" fill="#b4c3d2"/>',
        '<path d="M600 195H720V225H820V195H1000V270H600Z" fill="#96a5b9" stroke="#333"/>',
        '<rect x="732" y="195" width="76" height="30" fill="#19a67b"/>',
        '<text x="600" y="300">4 mm groove width × 2 mm deep</text>',
        '<text x="600" y="325">TPU: 2.8 mm wide × 2.4 mm free height</text>',
        '<text x="600" y="350">16.7% nominal squeeze; rigid lands limit travel</text>',
        '<text x="600" y="390">Continuous loop for every lid and interface</text>',
        '<text x="600" y="415">Blind insert pockets outside seal perimeter</text>',
        '<text x="600" y="440">Solid lids with 10 mm overlapping rain skirts</text>',
        '<text x="600" y="485">Orange interfaces in GLB are leak-test blanks.</text>',
        '<text x="600" y="510">Replace with engineered shaft / thermal plates.</text>',
        '<text x="600" y="535">External sealed cable links use selected glands.</text>',
        '<text x="30" y="650">Body assembly: 490 × 490 mm. Lids: 498 × 498 mm. Max closed height 226 mm; excludes drums and hardware.</text>',
        '<text x="30" y="680">Design intent: sustained rain/splash. No claimed IP rating. Validate seals, printed walls and cooling before service.</text>',
        '<text x="30" y="710">Print the small seal-test kit first. Confirm TPU hardness and saltwater exposure before finalising materials.</text>',
        '</svg>']
(OUT/'layout.svg').write_text('\n'.join(svg))
with zipfile.ZipFile(OUT.parent/'outdoor-enclosure-development.zip','w',zipfile.ZIP_DEFLATED) as archive:
    for path in sorted(OUT.rglob('*')):
        if path.is_file() and '__pycache__' not in path.parts:
            archive.write(path,path.relative_to(OUT.parent))
print(json.dumps(report,indent=2))
