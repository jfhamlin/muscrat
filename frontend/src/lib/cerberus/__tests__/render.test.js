import Cerberus from '..';

import createContext from 'gl';
import { createCanvas } from 'canvas';

const newGlContext = (width, height) => {
  const context = createContext(width, height);

  // headless-gl only supports WebGL 1. However, with a few hacks, we
  // can make THREE work reasonably well with headless-gl.
  // This is far from perfect. Even THREE's default shaders won't
  // compile since they use WebGL 2. However, this is irrelevant for 99%
  // of my tests.
  // Until headless-gl supports WebGL 2, tests that require shader
  // compilation, visual output, or WebGL data will need to use a browser
  // environment via Puppeteer.

  context.texImage3D = () => {};
  context.createVertexArray = () => {};
  context.bindVertexArray = () => {};
  context.deleteVertexArray = () => {};

  return context;
}

const newCanvas = (width, height) => {
  const canvas = createCanvas(width, height);
  canvas.addEventListener = (type, handler) => {
    canvas['on' + type] = handler.bind(canvas);
  };
  canvas.removeEventListener = (type) => {
    canvas['on' + type] = null;
  };
  canvas.style = {};
  return canvas;
};

describe('render', () => {
  it('should allocate no WebGL resources if nothing is rendered', async () => {

    const canvas = newCanvas(1, 1);
    let glContext = newGlContext(1, 1);

    let check = false;

    // No WebGL resources should be allocated. spy on the WebGL context to make sure.
    glContext = new Proxy(glContext, {
      get: (target, prop) => {
        if (check) {
          if (prop === 'createBuffer' ||
              prop === 'createTexture' ||
              prop === 'createFramebuffer' ||
              prop === 'createRenderbuffer' ||
              prop === 'createShader' ||
              prop === 'createProgram' ||
              prop === 'createVertexArray') {
            throw new Error(`WebGL resource allocation detected: ${prop}`);
          }
        }
        return target[prop];
      }
    });

    jest.spyOn(window, 'requestAnimationFrame').mockImplementation(cb => cb());

    const cerb = new Cerberus({
      canvas,
      glContext,
    });

    check = true;

    cerb.scene().add(
      cerb.sphere(1, 1, 1),
      cerb.pointLight(0xffffff),
    );

    setTimeout(() => {
      window.requestAnimationFrame.mockRestore();
    }, 0);
  });
});
