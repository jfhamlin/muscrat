/* Cerberus is a simple, functional library for live coding 3D
 * graphics in the browser. It is designed to be easy to use
 * and to provide a simple interface for creating 3D scenes.
 */

// three.js
import * as THREE from 'three';

const TWO_PI = 2 * Math.PI;

// MeshBasicMaterial

const DEFAULT_MATERIAL = new THREE.MeshLambertMaterial({
  color: 0xf0f0f0,
});

const getConstructorArgs = (cls) => {
  const ctorStr = cls.prototype.constructor.toString();
  const args = ctorStr.match(/\(([^)]*)\)/)[1]
    .split(',')
    .map((arg) => arg.trim())
    .filter(arg => arg !== '');
  // create a map from arg name to default value (undefined if none)
  const argMap = {};
  args.forEach((arg) => {
    const parts = arg.split('=').map((part) => part.trim());
    if (parts.length === 1) {
      argMap[parts[0]] = undefined;
    } else {
      // note that some default values will fail to eval
      argMap[parts[0]] = eval(parts[1]);
    }
  });
  return argMap;
};

class Cerberus {
  constructor(opts = {}) {
    const canvas = opts.canvas || document.createElement('canvas');
    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;

    this._canvas = canvas;
    this._renderer = new THREE.WebGLRenderer({
      canvas,
      antialias: true,
    });
    this._renderer.setSize(canvas.width, canvas.height);

    this._synthTime = 0; // time in milliseconds
    this._lastTime = undefined; // last render timestamp

    this._scenes = [];
    this._renderingScenes = [];

    const classNameToFnName = (className, suffix) => {
      const fnName = className.charAt(0).toLowerCase() + className.slice(1);
      return fnName.substring(0, fnName.length - suffix.length);
    };

    for (const key in THREE) {
      // add methods for all geometry types, by reflection
      if (key.endsWith('Geometry')) {
        const cls = THREE[key];
        // trim off 'Geometry' from key
        let fnName = key.substring(0, key.length - 8);
        fnName = fnName.charAt(0).toLowerCase() + fnName.slice(1);
        let args;
        try {
          args = getConstructorArgs(cls);
        } catch (e) {
          // skip if we can't get constructor args
          continue
        }
        this[fnName] = (...ctorArgs) => {
          const geometry = new cls(...ctorArgs);
          const mesh = new Object3D(new THREE.Mesh(geometry, DEFAULT_MATERIAL));
          return mesh;
        };
      }

      // add methods for all light types, by reflection
      if (key.endsWith('Light')) {
        const cls = THREE[key];
        // check that it inherits from Light
        if (cls.prototype instanceof THREE.Light) {
          const fnName = classNameToFnName(key, '');
          let args;
          try {
            args = getConstructorArgs(cls);
          } catch (e) {
            // skip if we can't get constructor args
            continue
          }
          this[fnName] = (...ctorArgs) => {
            const light = new Object3D(new cls(...ctorArgs));
            return light;
          }
        }
      }
    }
  }

  getCanvas() {
    return this._canvas;
  }

  setResolution(width, height) {
    this._canvas.width = width;
    this._canvas.height = height;
    this._renderer.setSize(width, height);
  }

  camera(fov = 75) {
    const camera = new THREE.PerspectiveCamera(
      fov, this._canvas.width / this._canvas.height,
      0.1, 1000);
    return new Object3D(camera);
  }

  scene() {
    const scene = new Scene(this);
    this._scenes.push(scene);
    return scene;
  }

  dispose() {
    console.log('disposing');
    this._scenes.forEach((scene) => scene.dispose());
    this._renderer.dispose();
  }

  _startRendering(scene) {
    if (this._renderingScenes.includes(scene)) {
      return;
    }

    this._renderingScenes.push(scene);
    if (this._renderingScenes.length > 1) {
      return;
    }

    const doRender = (now) => {
      if (this._renderingScenes.length === 0) {
        return;
      }

      if (!this._lastTime) {
        this._lastTime = now;
      } else {
        this._synthTime += now - this._lastTime;
        this._lastTime = now;
      }

      // tick in seconds!
      const timeSecs = this._synthTime / 1000;
      for (const scene of this._renderingScenes) {
        scene.tick(timeSecs);
        scene._show(); // need to decouple this from the single canvas
      }

      if (this._renderingScenes.length > 0) {
        requestAnimationFrame(doRender);
      }
    }

    requestAnimationFrame(doRender);
  }

  _stopRendering(scene) {
    const idx = this._renderingScenes.indexOf(scene);
    if (idx !== -1) {
      this._renderingScenes.splice(idx, 1);
    }
  }
}

class Scene {
  constructor(cerb) {
    this._cerb = cerb;
    this._scene = new THREE.Scene();
    this._objects = [];

    this._camera = undefined;
  }

  add(/* any number of Object3D instances */) {
    for (const obj of arguments) {
      this._objects.push(obj);
      this._scene.add(obj._obj)
    }
    return this;
  }

  tick(time) {
    if (this._camera) {
      this._camera.tick(time);
    }

    for (const obj of this._objects) {
      obj.tick(time);
    }
  }

  render(camera) {
    if (this._camera) {
      return;
    }

    if (!camera) {
      camera = this._cerb.camera();
      camera._obj.position.z = 5;
    }

    this._camera = camera;

    this._cerb._startRendering(this);
  }

  _show() {
    this._cerb._renderer.render(this._scene, this._camera._obj);
  }

  stop() {
    this._cerb._stopRendering(this);
  }

  dispose() {
    this.stop();

    this._scene.dispose();
  }
}

class Object3D {
  constructor(obj, tick) {
    this._obj = obj;
    this.tick = tick ?? (() => {});

    for (const fn of funcs) {
      this[fn.name] = (...args) => {
        const parent = this;
        const obj = this._obj.clone();
        if (fn.type === 'material' && obj.material) {
          obj.material = obj.material.clone();
        }
        let lastTickTime;

        const tick = (time) => {
          if (time === lastTickTime) {
            return;
          }
          lastTickTime = time;
          // make sure parent is updated first
          parent.tick(time);

          // update obj from parent
          obj.position.copy(parent._obj.position);
          obj.rotation.copy(parent._obj.rotation);
          obj.scale.copy(parent._obj.scale);

          // update obj material from parent if applicable
          if (obj.material) {
            obj.material.color.copy(parent._obj.material.color);
          }

          const vals = args.map((arg) => {
            if (typeof arg === 'function') {
              return arg({ time });
            }
            return arg;
          });
          fn.apply(time, obj, ...vals);
        };
        return new Object3D(obj, tick);
      };
    }
  }

  dispose() {}
}

const funcs = [{
  name: 'translate',
  type: 'xform',
  apply: (time, obj,
          x = 0, y = 0, z = 0,
          speedX = 0, speedY = 0, speedZ = 0) => {
            obj.position.set(
              x + speedX * time,
              y + speedY * time,
              z + speedZ * time,
            );
          },
}, {
  name: 'color',
  type: 'material',
  apply: (time, obj, r = 1, g = 1, b = 1) => {
    if (obj.isLight) {
      obj.color = new THREE.Color(r, g, b);
    } else {
      obj.material.color = new THREE.Color(r, g, b);
    }
  },
}, {
  name: 'scale',
  type: 'xform',
  apply: (time, obj,
          amount = 1.5, xMult = 1, yMult = 1, zMult = 1) => {
            obj.scale.set(
              amount * xMult,
              amount * yMult,
              amount * zMult,
            );
          },
}, {
  name: 'rotate',
  type: 'xform',
  apply: (time, obj,
          x = 0, y = 0, z = 0,
          speedX = 0, speedY = 0, speedZ = 0) => {
            const xRad = TWO_PI * x;
            const yRad = TWO_PI * y;
            const zRad = TWO_PI * z;

            const pitch = xRad + TWO_PI * speedX * time;
            const yaw = yRad + TWO_PI * speedY * time;
            const roll = zRad + TWO_PI * speedZ * time;

            const euler = new THREE.Euler(pitch, yaw, roll, 'XYZ');
            const quaternion = new THREE.Quaternion().setFromEuler(euler);

            obj.position.applyQuaternion(quaternion);
            obj.quaternion.premultiply(quaternion);
          },
}, {
  name: 'lookAt',
  type: 'xform',
  apply: (time, obj, x = 0, y = 0, z = 0) => {
    obj.lookAt(new THREE.Vector3(x, y, z));
  },
}];

export default Cerberus;
