"""Two-container outdoor concept v3. STL mm, GLB metres. No vendor CAD implied."""
from pathlib import Path
import json
import zipfile
import manifold3d as m
import numpy as np
import trimesh

OUT = Path(__file__).parent
STL = OUT/'stl'
STL.mkdir(exist_ok=True)
parts, report = {}, {}
W = 230
BASE_H = 180
INSERT_D = 5.8
TOP_BOLTS = [(7,7),(115,7),(223,7),(223,115),(223,223),(115,223),(7,223),(7,115)]
JOIN_BOLTS = [(28,28),(115,28),(202,28),(202,115),(202,202),(115,202),(28,202),(28,115)]

def box(size,p=(0,0,0)):
    return m.Manifold.cube(size).translate(p)

def ring(outer,inner,h,p=(0,0,0)):
    d=(outer-inner)/2
    return (box((outer,outer,h))-box((inner,inner,h+2),(d,d,-1))).translate(p)

def cyl(r,h,p=(0,0,0)):
    return m.Manifold.cylinder(h,r,r,64).translate(p)

def mesh(s):
    a=s.to_mesh()
    return trimesh.Trimesh(np.asarray(a.vert_properties)[:,:3],np.asarray(a.tri_verts),process=True)

def save(name,s,qty,material='rigid'):
    assert s.status()==m.Error.NoError and len(s.decompose())==1,name
    tri=mesh(s)
    assert tri.is_volume and tri.is_watertight,name
    assert max(tri.extents)<=240.001 and tri.bounds[0,2]>=-.001,name
    tri.export(STL/f'{name}.stl')
    check=trimesh.load_mesh(STL/f'{name}.stl')
    assert check.is_volume and check.is_watertight,name
    parts[name]=s
    report[name]=dict(quantity=qty,material=material,bounds_mm=tri.extents.round(3).tolist(),
                     volume_cm3=round(tri.volume/1000,2))

def shell(h,base):
    s=box((230,230,h))-box((218,218,h+2),(6,6,8 if base else -1))
    # Wide base mating rim has internal bolts, accessible through the upper rim.
    s += ring(230,158 if base else 186,12,(0,0,h-12))
    s -= ring(202,194,2.1,(14,14,h-2))
    for x,y in JOIN_BOLTS if base else TOP_BOLTS:
        s -= cyl(INSERT_D/2,6.1,(x,y,h-6))
    if base:
        for x,y in ((30,30),(200,30),(30,200),(200,200)):
            s -= cyl(INSERT_D/2,5,(x,y,0))
        for x in (20,115,210):
            for y in (25,115,205):
                s += cyl(6,6,(x,y,8))
                s -= cyl(1.7,5.1,(x,y,9))
    else:
        s += ring(230,158,12)
        for x,y in JOIN_BOLTS:
            s -= cyl(2.25,14,(x,y,-1))
    return s

def port_transform(s,side,cx,cz):
    # Port helper is local X/Z around (0,0), penetrates inward along +Y.
    s=s.translate((cx,0,cz))
    if side=='rear':
        return s.rotate((0,0,180)).translate((230,230,0))
    if side=='right':
        return s.rotate((0,0,90)).translate((230,0,0))
    if side=='left':
        return s.rotate((0,0,-90)).translate((0,230,0))
    return s

def port(s,side,cx,cz,opening,groove,cap,boltpitch):
    pad=box((cap+4,12,cap+4),(-(cap+4)/2,0,-(cap+4)/2))
    cut=box((opening,14,opening),(-opening/2,-1,-opening/2))
    g=ring(groove,groove-8,2.1).rotate((-90,0,0)).translate((-groove/2,-.1,groove/2))
    holes=m.Manifold()
    for x,z in bolt_points(boltpitch):
        holes += cyl(INSERT_D/2,6).rotate((-90,0,0)).translate((x,0,z))
    return (s+port_transform(pad,side,cx,cz))-port_transform(cut+g+holes,side,cx,cz)

def bolt_points(p):
    r=p/2
    return [(-r,-r),(0,-r),(r,-r),(r,0),(r,r),(0,r),(-r,r),(-r,0)]

def cap(size,pitch):
    s=box((size,size,6))
    for x,y in bolt_points(pitch):
        s -= cyl(2.25,8,(x+size/2,y+size/2,-1))
    return s

# Separate front shaft interfaces; 120 mm pitch for the two 110 mm demo drums.
winch_base=shell(BASE_H,True)
for x in (55,175):
    winch_base=port(winch_base,'front',x,65,70,82,104,92)
save('winch-base',winch_base,1)
winch_riser=port(shell(160,False),'rear',115,70,100,116,136,124)
winch_riser=port(winch_riser,'right',115,70,40,56,76,64)
save('winch-riser',winch_riser,1)

power_base=port(shell(BASE_H,True),'rear',115,90,100,116,136,124)
save('power-compute-base',power_base,1)
power_riser=port(shell(200,False),'rear',115,115,100,116,136,124)
power_riser=port(power_riser,'left',115,70,40,56,76,64)
save('power-compute-riser',power_riser,1)

lid=box((238,238,6))+ring(238,231,10,(0,0,6))
for x,y in TOP_BOLTS:
    lid -= cyl(2.25,8,(x+4,y+4,-1))
save('rain-skirt-lid',lid,2)
save('body-and-lid-gasket-tpu',ring(200.8,195.2,2.4),4,'TPU')
for name,size,pitch,g,qty in (('shaft',104,92,82,2),('thermal',136,124,116,3),('hybrid',76,64,56,2)):
    save(f'{name}-interface-blank',cap(size,pitch),qty)
    save(f'{name}-interface-gasket-tpu',ring(g-1.2,g-6.8,2.4),qty,'TPU')

# Removable dry-side mounting blanks; final holes match the actual equipment.
save('equipment-mounting-blank',box((150,150,4)),2)

components = {
 'winch':[
    ('motor-gearbox-A',(26.5,6,36.5),(57,195,57),'#e8a548'),
    ('motor-gearbox-B',(146.5,6,36.5),(57,195,57),'#e8a548'),
    ('driver-A',(56,80,145),(25.5,79.5,118),'#a882ce'),
    ('driver-B',(148,80,145),(25.5,79.5,118),'#a882ce'),
    ('DDR-15L-24',(25,154,210),(17.5,54.5,90),'#69b68b'),
    ('brake-relays-K1-K2-reservation',(185,128.5,210),(12.4,80,94),'#9bc4a6'),
 ],
 'power-compute':[
    ('NDR-240-48',(20,95,35),(63,113.5,125.2),'#69b68b'),
    ('Jetson-with-connectors-reservation',(12,30,265),(145,130,85),'#53b1c7'),
    ('DDR-60L-12',(169,150,261),(52.5,54.5,90),'#8cc49b'),
    ('USB-RS485-reservation',(25,180,265),(45,35,25),'#a882ce'),
 ]}

def installed_cap(name,side,cx,cz):
    s=parts[f'{name}-interface-blank']
    size=report[f'{name}-interface-blank']['bounds_mm'][0]
    s=s.rotate((90,0,0)).translate((-size/2,0,-size/2))
    return port_transform(s,side,cx,cz)

assemblies={}
for kind,height in (('winch',340),('power-compute',380)):
    items=[('base',parts[f'{kind}-base']),('riser',parts[f'{kind}-riser'].translate((0,0,180))),
           ('lid',lid.rotate((180,0,0)).translate((-4,234,height+6)))]
    if kind=='winch':
        items += [(f'shaft-cap-{x}',installed_cap('shaft','front',x,65)) for x in (55,175)]
        items += [('thermal-cap',installed_cap('thermal','rear',115,250)),
                  ('hybrid-cap',installed_cap('hybrid','right',115,250))]
    else:
        items += [('lower-thermal-cap',installed_cap('thermal','rear',115,90)),
                  ('upper-thermal-cap',installed_cap('thermal','rear',115,295)),
                  ('hybrid-cap',installed_cap('hybrid','left',115,250))]
    for i,(n,s) in enumerate(items):
        for other,t in items[:i]:
            assert (s ^ t).volume()<.001,(kind,n,other)
    for n,p,size,_ in components[kind]:
        for other,s in items:
            assert (box(size,p) ^ s).volume()<.001,(kind,n,other)
    for i,(n,p,size,_) in enumerate(components[kind]):
        for other,q,extent,_ in components[kind][:i]:
            assert (box(size,p) ^ box(extent,q)).volume()<.001,(kind,n,other)
    # Gasket groove fit at both joints; compressed-height location surrogate.
    for z in (178,height-2):
        seal=ring(200.8,195.2,2,(14.6,14.6,z))
        for n,s in items:
            assert (seal ^ s).volume()<.001,(kind,'gasket',n)
        items.append((f'gasket-{z}',seal))
    # Tool access to internal join screws through the narrower top opening.
    for x,y in JOIN_BOLTS:
        tool=cyl(2.2,height-191,(x,y,193))
        assert (tool ^ items[1][1]).volume()<.001,(kind,'join screw access')
    assemblies[kind]=items

assert 145-93.5>=50 and 328-263>=50 # drive lower/upper space, below rim
assert 56-6>=50 and 224-173.5>=50 and 148-81.5>=30
assert 261-160.2>=100 # PSU to upper heat sources, minimum separation
assert 2.8*2.4<4*2 # gasket displacement fits gland

scene=trimesh.Scene()
def add(s,name,color,offset=(0,0,0)):
    tri=mesh(s.translate(offset))
    tri.visual.vertex_colors=color
    scene.add_geometry(tri,node_name=name)

for kind,offset in (('winch',(0,0,0)),('power-compute',(350,0,0))):
    for n,s in assemblies[kind]:
        color=[35,180,130,255] if n.startswith('gasket') else [165,180,200,80]
        if 'cap' in n:
            color=[215,155,70,180]
        add(s,f'{kind}-{n}',color,offset)
    for n,p,size,c in components[kind]:
        add(box(size,p),f'{kind}-{n}',trimesh.visual.color.hex_to_rgba(c),offset)
# External drum locations are illustrative pending the shaft extension/seal stack.
for x in (55,175):
    add(cyl(55,50).rotate((90,0,0)).translate((x,-30,65)),f'drum-envelope-{x}',[235,160,50,220])
# One continuous illustrative umbilical, not a manufactured cable drawing.
points=[(236,115,250),(265,115,250),(280,115,220),(300,115,220),(315,115,250),(344,115,250)]
for i in range(len(points)-1):
    t=trimesh.creation.cylinder(radius=5,segment=np.array(points[i:i+2]),sections=24)
    t.visual.vertex_colors=[40,45,50,255]
    scene.add_geometry(t,node_name=f'single-hybrid-cable-{i}')
scene.apply_scale(.001)
scene.export(OUT/'assembly.glb')
(OUT/'component-layout.json').write_text(json.dumps(components,indent=2)+'\n')
(OUT/'validation.json').write_text(json.dumps(dict(parts=report,
    checks=['connected watertight positive-volume print meshes','exported STL reload',
        'all print extents <=240 mm','assembly part non-interference','component envelope non-interference',
        'main gasket groove fit','join screw tool paths','selected driver and PSU spatial clearances'],
    not_validated=['shaft cartridge/hub stack','component mounting and cable bends',
        'thermal performance','physical water sealing','hybrid connector selection']),indent=2)+'\n')

svg=['<svg xmlns="http://www.w3.org/2000/svg" width="1100" height="850" viewBox="0 0 1100 850">',
 '<style>text{font-family:Arial;font-size:15px}.title{font-size:25px;font-weight:bold}</style>',
 '<rect width="1100" height="850" fill="white"/>',
 '<text x="30" y="35" class="title">DepthPilot iteration 3 — two containers, one hybrid cable</text>',
 '<text x="30" y="65">Front elevation, component envelopes. Drums outside. Dimensions in mm; source/GLB include the full layout.</text>']
for kind,ox,h in (('winch',60,340),('power-compute',600,380)):
    oy=510-h
    svg += [f'<rect x="{ox}" y="{oy}" width="230" height="{h}" fill="#edf1f5" stroke="#333"/>',
            f'<line x1="{ox}" x2="{ox+230}" y1="330" y2="330" stroke="#159879" stroke-dasharray="6 4"/>',
            f'<text x="{ox}" y="{oy-14}">{kind}: 230 W × 230 D × {h+6} H</text>']
    for i,(n,(x,y,z),(a,b,c),color) in enumerate(components[kind],1):
        svg += [f'<rect x="{ox+x}" y="{510-z-c}" width="{a}" height="{c}" fill="{color}" stroke="#333"/>',
                f'<text x="{ox+x+3}" y="{510-z-c+16}">{i}</text>']
    for i,(n,*_) in enumerate(components[kind],1):
        svg += [f'<text x="{ox}" y="{550+22*(i-1)}">{i}. {n}</text>']
for x in (55,175):
    svg += [f'<circle cx="{60+x}" cy="445" r="55" fill="none" stroke="#b86e16" stroke-dasharray="5 3"/>']
svg += ['<path d="M290 260H365Q390 260 390 300Q390 330 420 330H455Q485 330 485 300Q485 260 510 260H600" fill="none" stroke="#333" stroke-width="8"/>',
 '<text x="330" y="225">ONE cable: 48 V + RS485</text>',
 '<text x="350" y="365">Power cores + shielded data pair</text>',
 '<text x="30" y="715">Continuous TPU gaskets at the body joint and lid. Shared interior in each container; no inter-box air passage.</text>',
 '<text x="30" y="742">Caps are development blanks for shaft, thermal and connector interfaces. Actual cartridges/sinks remain to be selected.</text>',
 '<text x="30" y="769">Each printed base/riser fits the printer; assembled boxes are taller. Common lids have a 238 × 238 mm footprint.</text>',
 '<text x="30" y="796">No IP rating or thermal qualification claimed. See README and ELECTRICAL.md before building.</text></svg>']
(OUT/'layout.svg').write_text('\n'.join(svg))
with zipfile.ZipFile(OUT.parent/'two-box-v3.zip','w',zipfile.ZIP_DEFLATED) as archive:
    for p in sorted(OUT.rglob('*')):
        if p.is_file() and '__pycache__' not in p.parts:
            archive.write(p,p.relative_to(OUT.parent))
print(json.dumps(report,indent=2))
