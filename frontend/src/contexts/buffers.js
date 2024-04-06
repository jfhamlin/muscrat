import { create } from "zustand";
import createContext from "zustand/context";

const {
  Provider: BuffersProvider,
  useStore: useBuffersStore,
} = createContext();

const DEFAULT_CODE = `(ns user
  (:require [mrat.core :refer :all]
    [mrat.scales :refer :all]
    [mrat.midi :refer :all]))

(play (sin 200))
`;

const createBuffersStore = () =>
  create((set, get) => ({
    selectedBufferName: null,
    buffers: {
      null: {
        fileName: '',
        content: DEFAULT_CODE,
        dirty: false,
      },
    },
    selectBuffer: (fileName) => set((state) => ({ selectedBufferName: fileName })),
    addBuffer: ({ fileName, content }) =>
      set((state) => ({
        selectedBufferName: fileName,
        buffers: {
          ...state.buffers,
          [fileName]: {
            fileName,
            content,
            dirty: false,
          },
        },
      })),
    updateBuffer: (fileName, content, dirty) => {
      set((state) => ({
        buffers: {
          ...state.buffers,
          [fileName]: {
            ...(state.buffers[fileName] ?? { fileName }),
            content,
            dirty: dirty ?? true,
          },
        },
      }));
    },
    cleanBuffer: (fileName, content) => {
      // if content is same as buffer content, then un-dirty the buffer
      set((state) => ({
        buffers: {
          ...state.buffers,
          [fileName]: {
            ...state.buffers[fileName],
            dirty: state.buffers[fileName]?.content !== content,
          },
        },
      }));
    },
  }));

export { BuffersProvider, useBuffersStore, createBuffersStore };
