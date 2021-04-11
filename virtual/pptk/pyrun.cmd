set PYTHONPATH=%~dp0benchmark/py;%~dp0;%PYTHONPATH%
call %HOME%\anaconda3\condabin\conda activate
python %~dp0toolkit.py %*

