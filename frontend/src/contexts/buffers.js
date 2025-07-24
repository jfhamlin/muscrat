import { create } from "zustand";
import createContext from "zustand/context";
import { SaveToTemp, SaveFile, PlayFile } from "../../bindings/github.com/jfhamlin/muscrat/muscratservice";

const {
  Provider: BuffersProvider,
  useStore: useBuffersStore,
} = createContext();

const DEFAULT_CODE = `(ns user
  (:use [mrat.core]))

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
    playBuffer: async (bufferKey) => {
      const buffer = get().buffers[bufferKey];
      if (!buffer) return;
      
      let fileToPlay = buffer.fileName;
      
      // If temp file, always update it with current content
      if (buffer.isTemp && fileToPlay) {
        try {
          // Write current content to the existing temp file
          await SaveFile(fileToPlay, buffer.content);
        } catch (err) {
          console.error("Failed to update temp file:", err);
          return;
        }
      } else if (!fileToPlay) {
        // If no fileName (unsaved buffer), create temp file
        try {
          fileToPlay = await SaveToTemp(buffer.content);
          // Update buffer with temp path for tracking
          set((state) => ({
            buffers: {
              ...state.buffers,
              [bufferKey]: {
                ...state.buffers[bufferKey],
                fileName: fileToPlay,
                isTemp: true,
              },
            },
          }));
        } catch (err) {
          console.error("Failed to save to temp:", err);
          return;
        }
      }
      
      await PlayFile(fileToPlay);
    },
  }));

export { BuffersProvider, useBuffersStore, createBuffersStore };
