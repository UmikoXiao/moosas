from ..thermal import idfGeometry, construction
from ..geometry.element import MoosasSpace
from eppy.modeleditor import IDF
import re


def writeIDF(model, idfTemplatePath: str, outputPath: str,idd='C:\EnergyPlusV23-1-0\Energy+.idd'):
    from ..models import MoosasModel
    model: MoosasModel = model
    print('IDF: initialization from IDF file...')
    # IDF.setiddname(r'C:\EnergyPlusV8-9-0\Energy+.idd')
    IDF.setiddname(idd)
    idf = IDF(idfTemplatePath)
    moElements = model.getAllFaces(dumpUseless=True)
    zTemplate = idfGeometry.ZoneTemplate(idf)
    hint = []
    zName = [obj['Name'] for obj in idf.idfobjects['Zone']]+[obj['Name'] for obj in idf.idfobjects['Space']]
    for key in idf.idfobjects:
        print(f"\rIDF: cleaning existing objects: {key}", end='')
        if len(idf.idfobjects[key]) > 0:
            for objName in idf.idfobjects[key][0].obj:
                if objName in zName:
                    hint.append(key)
                    break
    hint +=zTemplate.objectHint+['Zone','WaterUse:Equipment','BuildingSurface:Detailed','FenestrationSurface:Detailed','Space']
    for h in hint:
        idf.idfobjects[h] = []
        print(f"\rIDF: cleaning existing objects: {h}", end='')
    print()
    for wi, wall in enumerate(moElements['MoosasWall']):
        print(f"\rIDF: encoding walls: {wi}/{len(moElements['MoosasWall'])}", end='')
        space = model.spaceIdDict[wall.space[0]]
        if not space.is_void():
            wallU, winU, SHGC = space.settings['zone_wallU'], space.settings['zone_winU'], space.settings['zone_win_SHGC']
            wallConstruction = zTemplate.getConstruction('opaque', wallU)
            windowConstruction = zTemplate.getConstruction('window', winU,SHGC)
            idfGeometry.createThermalSurface(idf,wall,'Wall',wallConstruction.params['Name'],windowConstruction.params['Name'])
    print()
    for fi, face in enumerate(moElements['MoosasFace']):
        print(f"\rIDF: encoding faces: {fi}/{len(moElements['MoosasFace'])}", end='')
        faceType = 'Floor'
        space = model.spaceIdDict[face.space[0]]
        if not space.is_void():
            if len(face.space)==1:
                if face in model.spaceIdDict[face.space[0]].ceiling.face:
                    faceType = 'Roof'
            wallU, winU, SHGC = space.settings['zone_wallU'], space.settings['zone_winU'], space.settings['zone_win_SHGC']
            wallConstruction = zTemplate.getConstruction('opaque', wallU)
            windowConstruction = zTemplate.getConstruction('window', winU,SHGC)
            idfGeometry.createThermalSurface(idf,face,faceType,wallConstruction.params['Name'],windowConstruction.params['Name'])
    print()
    for si, space in enumerate(model.spaceList):
        print(f"\rIDF: encoding zones: {si}/{len(model.spaceList)}", end='')
        if space.is_void():
            print('***Warring: EnergyPlus do not support void space')
        else:
            zTemplate.appliedToZone(space)
    idf.save(outputPath)
    print()