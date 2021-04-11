from gobench import *

import base64
import tempfile
import pandas as pd
import matplotlib.pyplot as plt
import matplotlib.ticker as mtick
import graphviz as gv
from IPython.display import Markdown, display, Image

def strip(f):
    rules = [
        ("github.com/spacemeshos/",""),
        ("github.com/sudachen/",""),
        ("sudachen.xyz/","")
    ]
    for r in rules:
        if f.startswith(r[0]):
            f = r[1] + f[len(r[0]):]
    return f


def plot_pprof(title, label, p):
    display(Markdown("# "+title))
    h = int(len(p.rows)*4.5/10+0.5)
    fig, ax = plt.subplots(nrows=1, ncols=4, figsize=(15,h), sharey='row')
    attr = ('cum','cum%','flat','flat%')
    for i in range(4):
        df = pd.DataFrame.from_records([(strip(x.function), x[attr[i]]) for x in p.rows],columns=(label, attr[i]))
        df.set_index([label], inplace=True)
        df.plot(kind='barh', ax=ax[i]).invert_yaxis()
    plt.tight_layout()
    plt.show()

def plot_pprof_image(title, image):
    display(Markdown("# "+title))
    S = base64.b64decode(image).decode()
    tfn = tempfile.mktemp(suffix='.png', prefix='pgraphviz-')
    display(Image(gv.Source(S, format='png', engine='dot').render(tfn)))

def report(benchmark_file):
    with open(benchmark_file) as f:
        r, ppf = load_results(f)
        ppfm = {i.label:i for i in ppf}
    if len(ppfm['top'].rows) > 0:
        plot_pprof("TOP Calls", "calls", ppfm['top'])
        plot_pprof_image("TOP Calls Grapth", ppfm['top'].image)
    if len(ppfm['alloc'].rows) > 0:
        plot_pprof("TOP Allocs", "calls", ppfm['alloc'])
        plot_pprof_image("TOP Allocs Grapth", ppfm['alloc'].image)

TEXT = r'''{
 "cells": [
  {
   "cell_type": "code",
   "execution_count": null,
   "metadata": {
    "collapsed": true
   },
   "outputs": [],
   "source": [
    "import toolkit\n",
    "toolkit.report('<BENCHMARK>')"
   ]
  }
 ],
 "metadata": {
  "kernelspec": {
   "display_name": "Python 3",
   "language": "python",
   "name": "python3"
  },
  "language_info": {
   "codemirror_mode": {
    "name": "ipython",
    "version": 3
   },
   "file_extension": ".py",
   "mimetype": "text/x-python",
   "name": "python",
   "nbconvert_exporter": "python",
   "pygments_lexer": "ipython3",
   "version": "3.6.3"
  }
 },
 "nbformat": 4,
 "nbformat_minor": 2
}
'''


def remove_style_scoped(body):
    while True:
        i = body.find("<style scoped>")
        if i >= 0:
            j = body.find("</style>", i)
            body = body[:i] + body[j+8:]
        else:
            return body


def remove_python_code(body):
    while True:
        i = body.find("```python")
        if i >= 0:
            j = body.find("```\n", i)
            body = body[:i] + body[j+4:]
        else:
            return body


if __name__ == '__main__':
    import os, io, nbformat, asyncio, sys, codecs
    if sys.version_info[0] == 3 and sys.version_info[1] >= 8 and sys.platform.startswith('win'):
        asyncio.set_event_loop_policy(asyncio.WindowsSelectorEventLoopPolicy())

    from traitlets.config import Config
    from nbconvert.preprocessors import ExecutePreprocessor
    c = Config()

    if 'html' in sys.argv:
        image_dir = ''
        from nbconvert import HTMLExporter
        me = HTMLExporter(config=c)
        output = "README.html"
        strip_body = lambda b: b
    else:
        image_dir = '../_img'
        c.ExtractOutputPreprocessor.output_filename_template = '_img/{unique_key}_{cell_index}_{index}{extension}'
        from nbconvert import MarkdownExporter
        me = MarkdownExporter(config=c)
        output = "../README.md"
        strip_body = lambda b: remove_python_code(remove_style_scoped(b))

    nb = nbformat.reads(TEXT.replace('<BENCHMARK>',sys.argv[1]),nbformat.NO_CONVERT)
    ep = ExecutePreprocessor(timeout=600, kernel_name='python3')
    ep.preprocess(nb, {'metadata': {'path': '.'}})
    (body, r) = me.from_notebook_node(nb)
    if image_dir and not os.path.isdir(image_dir):
        os.mkdir(image_dir)
    if image_dir:
        for n, d in r.get('outputs',{}).items():
            with open('../'+n,"wb") as img:
                img.write(d)
    with codecs.open(output,"w","utf-8") as f:
        f.write(strip_body(body))
    f.close()
