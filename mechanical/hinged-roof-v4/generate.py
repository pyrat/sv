"""Hinged winch hood v4; mm STL, metres GLB. Metal parts have print-fit templates."""
from pathlib import Path
import json
import math
import zipfile
import manifold3d as m
import numpy as np
import trimesh

OUT=Path(__file__).parent
STL=OUT/'stl'
STL.mkdir(exist_ok=True)
parts={}
report={}
AXIS=(0,246,180)
HEIGHT=390
TOP=[(7,7),(115,7),(223,7),(223,115),(223,223),(115,223),(7,223),(7,115)]

def box(size,p=(0,0,0)):
    return m.Manifold.cube(size).translate(p)

def cyl(r,h,p=(0,0,0)):
    return m.Manifold.cylinder(h,r,r,48).translate(p)

def ring(o,i,h,p=(0,0,0)):
    return (box((o,o,h))-box((i,i,h+2),((o-i)/2,(o-i)/2,-1))).translate(p)

def mesh(s):
    a=s.to_mesh()
    return trimesh.Trimesh(np.asarray(a.vert_properties)[:,:3],np.asarray(a.tri_verts),process=True)

def save(n,s,q=1,material='ASA/PETG prototype'):
    t=mesh(s)
    assert len(s.decompose())==1 and t.is_volume and t.is_watertight,n
    assert max(t.extents)<=240.001 and t.bounds[0,2]>=-.001,n
    t.export(STL/f'{n}.stl')
    check=trimesh.load_mesh(STL/f'{n}.stl')
    assert check.is_volume and check.is_watertight,n
    parts[n]=s
    report[n]=dict(quantity=q,material=material,extents_mm=t.extents.round(3).tolist())

def swing(s,angle):
    return s.translate(tuple(-v for v in AXIS)).rotate((-angle,0,0)).translate(AXIS)

def yhole(r,h,x,y,z):
    return cyl(r,h).rotate((-90,0,0)).translate((x,y,z))

def xhole(r,h,x,y,z):
    return cyl(r,h).rotate((0,90,0)).translate((x,y,z))

def boltpoints(p):
    r=p/2
    return [(-r,-r),(0,-r),(r,-r),(r,0),(r,r),(0,r),(-r,r),(-r,0)]

def port(s,cx,cz,opening,g,cap,pitch,side='front'):
    def position(a):
        a=a.translate((cx,0,cz))
        if side=='right':
            a=a.rotate((0,0,90)).translate((230,0,0))
        return a
    pad=box((cap+4,12,cap+4),(-(cap+4)/2,0,-(cap+4)/2))
    cut=box((opening,14,opening),(-opening/2,-1,-opening/2))
    cut+=ring(g,g-8,2.1).rotate((-90,0,0)).translate((-g/2,-.1,g/2))
    for x,z in boltpoints(pitch):
        cut+=yhole(2.9,6,x,0,z)
    return s+position(pad)-position(cut)

def blank(size,pitch):
    a=box((size,size,6))
    for x,y in boltpoints(pitch):
        a-=cyl(2.25,8,(x+size/2,y+size/2,-1))
    return a

# Wide permanent rim below the hinged hood; no inaccessible internal join bolts.
base=box((230,230,180))-box((218,218,174),(6,6,8))
base+=ring(230,186,12,(0,0,168))
base-=ring(202,194,2.1,(14,14,178))
for x in (55,175):
    base=port(base,x,65,70,82,104,92)
# Fixed-side cable interface sits on base, clear of the motor reservation.
base=port(base,120,135,40,56,76,64,'right')
for x in (30,200):
    for y in (30,200):
        base-=cyl(2.9,5,(x,y,0))

hood=box((230,230,210))-box((218,218,212),(6,6,-1))
hood+=ring(230,186,12)+ring(230,186,12,(0,0,198))
hood-=ring(202,194,2.1,(14,14,208))
for x,y in TOP:
    hood-=cyl(2.9,6.1,(x,y,204))

# Hinge bracket mount pads: four blind M4 pockets per leaf, rear face Y=230.
for start in (26,160):
    for s_is_base,z0 in ((True,140),(False,194)):
        z=z0 if s_is_base else z0-180
        pad=box((48,12,30),(start-2,218,z-2))
        for x in (start+6,start+38):
            for dz in (5,21):
                pad-=yhole(2.9,8.1,x,222,z+dz)
        if s_is_base:
            base+=pad
            for x in (start+6,start+38):
                for dz in (5,21):
                    base-=yhole(2.9,8.1,x,222,z+dz)
        else:
            hood+=pad
            for x in (start+6,start+38):
                for dz in (5,21):
                    hood-=yhole(2.9,8.1,x,222,z+dz)

# Two front compression-latch mounting pads. Hole patterns selected after hardware.
for x in (18,188):
    base+=box((24,8,26),(x,4,143))
    hood+=box((24,8,26),(x,4,12))
# Blind M4 clevis mounting pads for the removable open-position prop.
for z,is_base in ((145,True),(30,False)):
    pad=box((12,36,36),(0,22,z-18))
    if is_base: base+=pad
    else: hood+=pad
    for y in (28,52):
        for dz in (-12,12):
            cut=xhole(2.9,8.1,-.1,y,z+dz)
            if is_base: base-=cut
            else: hood-=cut
# Four external screw-clamp stations keep all closing screws outside the gasket.
for side in ('left','right'):
    for y in (70,185):
        x=0 if side=='left' else 218
        base+=box((12,24,30),(x,y-12,150))
        hood+=box((12,24,30),(x,y-12,0))
        start=-.1 if side=='left' else 222
        for z in (155,165): base-=xhole(2.9,8.1,start,y,z)
        for z in (15,25): hood-=xhole(2.9,8.1,start,y,z)
for x in (20,210):
    for y in (20,210):
        base+=cyl(7,6,(x,y,8))
        base-=cyl(2.9,5.1,(x,y,9))
save('winch-base',base)
save('hinged-hood',hood)

lid=box((238,238,6))+ring(238,231,10,(0,0,6))
for x,y in TOP:
    lid-=cyl(2.25,8,(x+4,y+4,-1))
carrier_holes=[(40,45),(190,45),(40,190),(190,190)]
for x,y in carrier_holes:
    lid+=cyl(7,5,(x+4,y+4,6))
    lid-=cyl(2.9,6,(x+4,y+4,6)) # opens at boss tip, 5 mm roof remains
save('roof',lid)
save('main-gasket-tpu',ring(200.8,195.2,2.4),2,'TPU; calibrate compression')
for name,size,pitch,g,qty in (('shaft',104,92,82,2),('hybrid',76,64,56,1)):
    save(f'{name}-blank',blank(size,pitch),qty)
    save(f'{name}-gasket-tpu',ring(g-1.2,g-6.8,2.4),qty,'TPU')

# Pin axis parallel to X. Two static ears surround one moving knuckle.
fixed=box((44,6,26),(0,230,140))
for x in (0,32):
    fixed+=box((12,18,22),(x,236,158))
    fixed+=xhole(8,12,x,246,180)
    fixed-=xhole(3.2,14,x-1,246,180)
moving=box((44,6,26),(0,230,194))
moving+=box((19,18,22),(12.5,236,180))
moving+=xhole(8,19,12.5,246,180)
moving-=xhole(3.2,21,11.5,246,180)
for x in (6,38):
    for dz in (5,21):
        fixed-=yhole(2.25,8,x,229,140+dz)
        moving-=yhole(2.25,8,x,229,194+dz)
# Print on mounting faces; metal fabrication templates for the final loaded hood.
save('hinge-fixed-fit-template',fixed.translate((0,-230,-140)).rotate((90,0,0)).translate((0,48,0)),2,'Fit template; final metal bracket')
save('hinge-moving-fit-template',moving.translate((0,-230,-172)).rotate((90,0,0)).translate((0,48,0)),2,'Fit template; final metal bracket')

clevis=box((6,32,32),(-6,24,129))+box((38,16,16),(-38,32,137))
clevis-=box((8,17,18),(-34,33,136))
clevis-=xhole(2.75,40,-39,40,145)
for y in (28,52):
    for z in (133,157):
        clevis-=xhole(2.25,8,-7,y,z)
save('prop-clevis-fit-template',clevis.translate((38,-24,-129)).rotate((0,-90,0)).translate((32,0,0)),2,'Fit template; final metal clevis')

clamp_fixed=box((6,20,30),(230,-10,150))+box((20,20,8),(230,-10,172))
clamp_moving=box((6,20,30),(230,-10,180))+box((20,20,8),(230,-10,180))
clamp_fixed-=cyl(1.65,7,(243,0,173)) # tap M4 in metal; pilot is not printed thread
clamp_moving-=cyl(2.25,10,(243,0,179))
for z in (155,165): clamp_fixed-=xhole(2.25,8,229,0,z)
for z in (195,205): clamp_moving-=xhole(2.25,8,229,0,z)
save('closing-clamp-base-fit-template',clamp_fixed.translate((-230,10,-150)),4,'Metal bracket, M4 tapped hole; fit template only')
save('closing-clamp-hood-fit-template',clamp_moving.translate((-230,10,-180)),4,'Metal bracket; fit template only')

# Metal carrier: top plate and two suspended vertical equipment plates.
top=box((170,172.5,3),(30,35,382))
for x,y in carrier_holes:
    top-=cyl(2.25,5,(x,y,381))
driver_plate=box((140,3,130),(45,144.5,205))
rail_plate=box((175,3,112),(28,204.5,248))
straps=[]
for x in (50,170):
    straps.append(box((8,3,47),(x,144.5,335)))
    straps.append(box((8,3,22),(x,204.5,360)))
save('carrier-top-fit-template',top.translate((-30,-35,-382)),1,'3 mm aluminium; printed fit template')
save('driver-panel-fit-template',driver_plate.translate((-45,-144.5,-205)).rotate((90,0,0)).translate((0,130,0)),1,'3 mm aluminium; printed fit template')
save('din-panel-fit-template',rail_plate.translate((-28,-204.5,-248)).rotate((90,0,0)).translate((0,112,0)),1,'3 mm aluminium; printed fit template')
save('driver-suspension-fit-template',box((8,47,3)),2,'Metal welded suspension; fit template')
save('din-suspension-fit-template',box((8,22,3)),2,'Metal welded suspension; fit template')
motor_base=box((210,206,3),(10,14,16))
for x in (20,210):
    for y in (20,210): motor_base-=cyl(2.25,5,(x,y,15))
save('motor-base-fit-template',motor_base.translate((-10,-14,-16)),1,'3 mm metal plate; match actual brackets/load')
spacer=cyl(5,2)-cyl(2.25,4,(0,0,-1))
save('motor-base-spacer',spacer,4)

components=[
 ('motor-A',(26.5,6,36.5),(57,195,57)),
 ('motor-B',(146.5,6,36.5),(57,195,57)),
 ('driver-A',(56,65,210),(25.5,79.5,118)),
 ('driver-B',(148,65,210),(25.5,79.5,118)),
 ('24V-converter',(35,150,260),(17.5,54.5,90)),
 ('brake-relays',(185,124.5,260),(12.4,80,94)),
]
stationary=[('base',base),('motor-base-plate',motor_base)]
for x in (20,210):
    for y in (20,210): stationary.append((f'base-spacer-{x}-{y}',spacer.translate((x,y,14))))
mobile=[('hood',hood.translate((0,0,180))),
 ('roof',lid.rotate((180,0,0)).translate((-4,234,396))),
 ('carrier-top',top),('driver-panel',driver_plate),('din-panel',rail_plate)]
mobile += [(f'suspension-{i}',s) for i,s in enumerate(straps)]
stationary.append(('fixed-prop-clevis',clevis))
mobile.append(('moving-prop-clevis',clevis.translate((0,0,65))))
for side in ('left','right'):
    for y in (70,185):
        for s,target,label in ((clamp_fixed,stationary,'base'),(clamp_moving,mobile,'hood')):
            if side=='left': s=s.rotate((0,0,180)).translate((230,0,0))
            target.append((f'{side}-closing-clamp-{label}-{y}',s.translate((0,y,0))))
for x in (26,160):
    stationary.append((f'fixed-hinge-{x}',fixed.translate((x,0,0))))
    mobile.append((f'moving-hinge-{x}',moving.translate((x,0,0))))
for n,p,size in components:
    (stationary if n.startswith('motor') else mobile).append((n,box(size,p)))
for x in (55,175):
    stationary.append((f'shaft-cap-{x}',parts['shaft-blank'].rotate((90,0,0)).translate((x-52,0,13))))
stationary.append(('hybrid-cap',parts['hybrid-blank'].rotate((90,0,90)).translate((230,82,97))))

# Pairwise closed assembly clearance (designed touching surfaces have zero volume).
allparts=stationary+mobile
for i,(n,s) in enumerate(allparts):
    for other,t in allparts[:i]:
        assert (s ^ t).volume()<.01,('closed collision',n,other,(s^t).volume())

# Test rigid moving assembly against fixed geometry at 1-degree intervals.
fixed_union=m.Manifold.batch_boolean([s for _,s in stationary],m.OpType.Add)
moving_union=m.Manifold.batch_boolean([s for _,s in mobile],m.OpType.Add)
for angle in range(91):
    volume=(swing(moving_union,angle)^fixed_union).volume()
    assert volume<.01,('opening collision',angle,volume)

seal=ring(200.8,195.2,2,(14.6,14.6,178))
roofseal=seal.translate((0,0,210))
assert (seal^fixed_union).volume()<.01
assert (seal^moving_union).volume()<.01
assert (roofseal^moving_union).volume()<.01

# Removable open-position metal prop: dimensioned endpoints, no long printed strut.
prop_a=np.array([-30.,40.,145.])
prop_b=np.array([-30.,276.,386.]) # hood local point (-30,40,210) after 90 degrees
prop_length=float(np.linalg.norm(prop_b-prop_a))
prop=box((6,12,prop_length),(-33,34,145))
for z in (145,145+prop_length):
    prop+=xhole(6,6,-33,40,z)
    prop-=xhole(2.75,8,-34,40,z)
prop=prop.translate(tuple(-v for v in prop_a)).rotate((-math.degrees(math.atan2(236,241)),0,0)).translate(tuple(prop_a))
assert (prop^fixed_union).volume()<.01,'Prop hits fixed assembly'
assert (prop^swing(moving_union,90)).volume()<.01,'Prop hits open hood'
mesh(prop).export(OUT/'metal-prop-assembly-reference.stl')

def scene(angle):
    sc=trimesh.Scene()
    for n,s in stationary+[(n,swing(s,angle)) for n,s in mobile]:
        tri=mesh(s)
        color=[160,180,200,95]
        if n.startswith('motor'): color=[230,165,60,255]
        if n.startswith('driver-') and n!='driver-panel': color=[165,120,210,255]
        if n in ('24V-converter','brake-relays'): color=[75,180,120,255]
        if 'panel' in n or 'carrier' in n or 'suspension' in n: color=[150,155,160,255]
        if 'hinge' in n: color=[90,100,115,255]
        tri.visual.vertex_colors=color
        sc.add_geometry(tri,node_name=n)
    for n,s in [('joint-gasket',seal),('roof-gasket',swing(roofseal,angle))]:
        tri=mesh(s); tri.visual.vertex_colors=[25,180,120,255]
        sc.add_geometry(tri,node_name=n)
    if angle==90:
        tri=mesh(prop)
        tri.visual.vertex_colors=[220,60,60,255]
        sc.add_geometry(tri,node_name='removable-metal-prop')
    sc.apply_scale(.001)
    sc.export(OUT/f'winch-{angle:02d}-degrees.glb')
for angle in (0,45,90): scene(angle)

openbounds=mesh(swing(moving_union,90)).bounds.round(2).tolist()
closedbounds=mesh(fixed_union+moving_union).bounds.round(2).tolist()
(OUT/'validation.json').write_text(json.dumps(dict(parts=report,
    closed_body_lid_mm=[238,238,396],hinge_axis_mm=AXIS,
    closed_model_bounds_mm=closedbounds,
    open_mobile_bounds_mm=openbounds,opening_samples_degrees=list(range(91)),
    opening_intersection_tolerance_mm3=.01,prop_pin_centres_mm=[prop_a.tolist(),prop_b.tolist()],
    prop_hole_pitch_mm=round(prop_length,2),
    checks=['single connected watertight STL parts','STL reload','print bounds <=240 mm',
      'closed pairwise interference','0..90 degree rigid swept samples','main gasket fit','open prop fit'],
    exclusions=['fasteners','wiring motion','elastic gasket contact',
      'hinge strength','thermal design','actual component mount patterns','water testing']),indent=2)+'\n')
(OUT/'component-layout.json').write_text(json.dumps(components,indent=2)+'\n')

# Side elevation uses Y horizontally, Z vertically; closed/open overlaid.
svg=['<svg xmlns="http://www.w3.org/2000/svg" width="1100" height="760" viewBox="0 0 1100 760">',
 '<style>text{font-family:Arial;font-size:15px}.title{font-size:24px;font-weight:bold}</style>',
 '<rect width="1100" height="760" fill="white"/>',
 '<text x="30" y="35" class="title">DepthPilot hinged roof v4 — complete upper hood lifts</text>',
 '<text x="30" y="65">Side elevation: front at left. Closed shell 238 W × 238 D × 396 H mm, excluding external hardware.</text>']
def polygon(s,color,opacity):
    # Geometry silhouette via convex hull of projected mesh vertices.
    pts=sorted(set(map(tuple,mesh(s).vertices[:,1:3])))
    def cross(o,a,b):
        return (a[0]-o[0])*(b[1]-o[1])-(a[1]-o[1])*(b[0]-o[0])
    lower=[]; upper=[]
    for seq,chain in ((pts,lower),(pts[::-1],upper)):
        for p in seq:
            while len(chain)>=2 and cross(chain[-2],chain[-1],p)<=0: chain.pop()
            chain.append(p)
    points=' '.join(f'{70+y:.2f},{610-z:.2f}' for y,z in lower[:-1]+upper[:-1])
    return f'<polygon points="{points}" fill="{color}" fill-opacity="{opacity}" stroke="{color}"/>'
for n,s in [('base',base),('closed',moving_union)]:
    svg.append(polygon(s,'#657a90',.15))
svg.append(polygon(swing(moving_union,90),'#35a383',.2))
for n,p,size in components:
    s=box(size,p)
    svg.append(polygon(s,'#df9d3d' if n.startswith('motor') else '#9e6fc2',.65))
    if not n.startswith('motor'): svg.append(polygon(swing(s,90),'#9e6fc2',.45))
svg += ['<circle cx="316" cy="430" r="5" fill="#d44"/>',
 '<text x="340" y="455">Hinge axis Y=246, Z=180</text>',
 '<text x="590" y="150">Rear hinge + 90° removable prop</text>',
 '<text x="590" y="180">Motor/gearbox pairs fixed to metal base</text>',
 '<text x="590" y="210">Drivers + converter + relays move with roof</text>',
 '<text x="590" y="240">Fixed cable entry on right side of base</text>',
 '<text x="590" y="270">TPU gasket at Z=180 closing joint</text>',
 '<text x="590" y="300">Upper shell is 230 × 230 × 210 for printing</text>',
 '<text x="30" y="665">Metal carrier and hinge templates included. Open/closed/45° GLBs show component reservations and mounting panels.</text>',
 '<text x="30" y="695">Rigid motion sampled every 1°. Wiring, hardware loads, thermal performance and water sealing require physical verification.</text>',
 '</svg>']
(OUT/'hinged-layout.svg').write_text('\n'.join(svg))
with zipfile.ZipFile(OUT.parent/'hinged-roof-v4.zip','w',zipfile.ZIP_DEFLATED) as archive:
    for p in sorted(OUT.rglob('*')):
        if p.is_file() and '__pycache__' not in p.parts:
            archive.write(p,p.relative_to(OUT.parent))
print(json.dumps(dict(print_parts=len(report),opening_samples=91,open_bounds=openbounds,prop_pitch=prop_length),indent=2))
