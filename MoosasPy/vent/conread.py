"""
    Support module for auto_contamx.
    Including all IO method and the execution and explanation of simread.exe
    It has not been added into the document and should never be used individually.
"""

import re
import numpy as np
from ..utils.tools import callCmd

AIR_DENSITY = 1.204


def exe_simread(simread_path, file_path, responseFile):
    """simread.exe"""
    # print(simread_path + ' ' + file_path+'<'+responseFile)
    callCmd([simread_path, file_path + '<' + responseFile])


def read_flowpath(file_path):
    """*.lfr files generated by simread.exe"""
    path = []
    with open(file_path, 'r') as f:
        path = f.readlines()
        path = [re.split('\t[ ]*', li.strip('\n'))[0:6] for li in path]
        # print(path)
    return np.array(path)[1:, -2:].astype(float)


def read_temp(file_path):
    """*.nfr files generated by simread.exe"""
    node = []
    with open(file_path, 'r') as f:
        node = f.readlines()
        node = [re.split('\t[ ]*', li.strip('\n'))[0:6] for li in node]
        print(node)

    return np.array(node)[1:, 3].T.astype(float) + 273.15


def read_topology(file_path):
    """construct path topology from *.prj file in array"""
    head = ''
    topo = ''
    rear = ''
    with open(file_path) as f:
        line = f.readline()
        head += line
        found = False
        while line:
            if re.search('flow paths', line) != None:
                found = True
                break
            line = f.readline()
            head += line
        if not found:
            print('error: result not found')

        head += f.readline()
        line = f.readline()

        while line != '-999\n':
            topo += line
            line = f.readline()

        while line:
            rear += line
            line = f.readline()
    topo = re.split('\n', topo)
    topo.pop()
    topo = np.array([re.split('\s+', li.strip('\n')) for li in topo])
    return topo[:, 3:5].astype(int)


def build_matrix(file_path):
    """
        build AFN matrix from *.lfr,*.nfr and *.prj files
        the result's unit is transformed from kg/s into m3/h
    """
    if file_path[-4:] != '.prj':
        raise Exception(f'Wrong *.prj path: {file_path}')
    lfrFile = file_path[:-4]+'.lfr'
    nfrFile = file_path[:-4]+'.nfr'
    zone_length = len(read_temp(nfrFile))
    topology = read_topology(file_path)
    airflow = read_flowpath(lfrFile)
    # [n+1,n+1]2d matrix, n room and 1 ambt
    matrix = [[0] * (zone_length) for x in range(zone_length)]
    for flow, mass in zip(topology, airflow):
        if flow[0] == -1:
            flow[0] = zone_length
        if flow[1] == -1:
            flow[1] = zone_length
        _from, _to = 0, 0
        if mass[0] > 0:
            _from += mass[0]
        else:
            _to -= mass[0]
        if mass[1] > 0:
            _from += mass[1]
        else:
            _to -= mass[1]
        matrix[flow[0] - 1][flow[1] - 1] += _from
        matrix[flow[1] - 1][flow[0] - 1] += _to

    return np.array(matrix) * 3600.0 / AIR_DENSITY


def read_file(path):
    """
    We need to change the temperature in the project file for each iteration.
    therefore we need to split the head, temperature definition, and rear part of the prj file.
    """
    head = ''
    temp = ''
    rear = ''
    with open(path) as f:
        line = f.readline()
        head += line
        found = False
        while line:
            if re.search('zones', line) != None:
                found = True
                break
            line = f.readline()
            head += line
        if not found:
            print('error: result not found')

        head += f.readline()
        line = f.readline()

        while line != '-999\n':
            temp += line
            line = f.readline()

        while line:
            rear += line
            line = f.readline()
    return head, temp, rear


def read_zone(path):
    """
    get the zone name and zone volume in the project file
    zone[:, 8] is the volume, zone[:, 11] is the zone name
    """
    head, zone, rear = read_file(path)
    zone = np.array([re.split(r'[ ]+', li) for li in zone.split('\n')[0:-1]])
    return zone[:, 8], zone[:, 11]


def write_file(path, head, temp, rear):
    """
    write the new header, temperature and rear into the project file.
    """
    with open(path, 'w+') as f:
        f.write(head)
        f.write(temp)
        f.write(rear)
    return path


def read_result(path):
    """
    legacy method to read result in the txt file exported by contamW.
    """
    with open(path) as f:
        for i in range(6):
            line = f.readline()
        result = []
        while line != '':
            line = [str(a).strip(' ') for a in line.strip('\n').split('\t')]
            result.append(line)
            line = f.readline()
        result[-1].append(0)
        result = np.array(result)
        return result[:, 1], result[:, -1].astype(float), result[:, 2:-1].astype(float)
