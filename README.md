# MOOSAS Ver. 0.8.0

This is a program for building performance anaylsis and optimization for
the building sketch design stage targeted on architects. Most of the detail settings and
geometrical representations transforming, which are always confusing to architects,
will be solved behind the interface. The core of MOOSAS is built on python,
the interface is built on javascript & html, and the extensions are built on
golong including *.epw transformation, *.skp/*.obj transformation, wind pressure prediction, etc.
Besides, a SketchUp Plugin version was embedded by src/ coded in ruby as an interface of MOOSAS

## 👀 instructions

[SketchUp instruction](MoosasPy/doc/Users Manual.pdf)  
[python module document](MoosasPy/doc/document.md)


## 🔧 python installation
following module would need if an external Interpreter was implemented (see requirement):  
```commandline
pip install -r requirement.txt
```
Besides, we have embedded a portable python 3.11 in the python folder.  
Please deploy the project by toSketchUp.py in order to get the stable environment.  
In this case, the [MoosasPy](python/Lib/MoosasPy) could be call directly in this portable python
```commandline
python toSketchUp.py
```
If you want to call MoosasPy with your own Interpreter, you should include the MoosasPy in your system path like
```python
import os
os.environ['PATH']+=os.path.abspath(r'python/Lib')
import MoosasPy
```
please do not move the MoosasPy module to elsewhere since the directories were lock with relative path
like /db or /data, etc.


## 🔧 SketchUp installation
### packing the plugin files into *.rbz file

The SketchUp interface has been embedded in \src  
Please deploy the project by toSketchUp.py in order to get the standalone environment.  
```commandline
python toSketchUp.py
```
If success, you would get a **moosas.rbz** on the root.  
\*.rbz could be recognized by SketchUp.  
### install the plugin into SketchUp
1. Prepare the latest version of SketchUp Software [later than SketchUp 2019](https://www.sketchup.com/)
2. Open the SketchUp, open Extensions => Extension Manager in the toolbar.
3. Click “Install Extension” and find the *.rbz
4. If users want to try the CFD simulation, please also install blueCFD 2017-2. We should provide the installation package, or users can download from: [Downloads · blueCFD-Core Project](https://bluecfd.github.io/Core/Downloads/#bluecfd-core-2017-2)
   * Please do not change the default settings of file location during the installation. The program call blueCFD ONLY FROM: C:\Program Files\blueCFD-Core-2017

### MOOSAS License is expired?
for the openaccess version of Moosas, you can extend the license as you want.  
Open the [MoosasLock file](src/MoosasLock.rb), find the following line:  
@exp_date = [2025,9,1]  


## 📖 Reference
Please cite the following article if the analysis module was implemented:  
Lin, B.; Chen, H.; Yu, Q.; Zhou, X.; Lv, S.; He, Q.; Li, Z. MOOSAS – A Systematic Solution for Multiple Objective Building Performance Optimization in the Early Design Stage. Building and Environment 2021, 200, 107929. https://doi.org/10.1016/j.buildenv.2021.107929.
Please cite the following article if the transformation module was implemented:  
Xiao, J.; Zhou, H.; Yang, S.; Zhang, D.; Lin, B. A CAD-BEM Geometry Transformation Method for Face-Based Primary Geometric Input Based on Closed Contour Recognition. Build. Simul. 2023. https://doi.org/10.1007/s12273-023-1081-6.


## 🤗 Credits and acknowledgements

Developed by Research team directed by **Prof. Borong Lin** from Key Laboratory of Eco Planning & Green Building, Ministry of Education, Tsinghua University.  
**For colaboration, Please contact:**  
linbr@tsinghua.edu.cn  
**If you have any technical problems, Please reach to:**  
junx026@gmail.com
