import { create } from "zustand";
import createContext from "zustand/context";
// @ts-ignore - Wails generated bindings
import { SaveToTemp, SaveFile, PlayFile } from "../../bindings/github.com/jfhamlin/muscrat/muscratservice";
import { BuffersState } from "../types";

const {
  Provider: BuffersProvider,
  useStore: useBuffersStore,
  // @ts-ignore - zustand context type issue
} = createContext();

const DEFAULT_CODE = `(ns user
  (:use [mrat.core]))

(play (sin 200))
`;

const createBuffersStore = () =>
  create<BuffersState>((set, get) => ({
    selectedBufferName: null,
    buffers: {
      'null': {
        fileName: '',
        content: DEFAULT_CODE,
        dirty: false,
      },
    } as any,
    selectBuffer: (fileName: string | null) => set(() => ({ selectedBufferName: fileName })),
    addBuffer: ({ fileName, content }: { fileName: string; content: string }) =>
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
    updateBuffer: (fileName: string | null, content: string, dirty?: boolean) => {
      const key = fileName ?? 'null';
      set((state) => ({
        buffers: {
          ...state.buffers,
          [key]: {
            ...(state.buffers[key] ?? { fileName }),
            content,
            dirty: dirty ?? true,
          },
        },
      }));
    },
    cleanBuffer: (fileName: string | null, content: string) => {
      // if content is same as buffer content, then un-dirty the buffer
      const key = fileName ?? 'null';
      set((state) => ({
        buffers: {
          ...state.buffers,
          [key]: {
            ...state.buffers[key],
            dirty: state.buffers[key]?.content !== content,
          },
        },
      }));
    },
    playBuffer: async (bufferKey: string | null) => {
      const key = bufferKey ?? 'null';
      const buffer = get().buffers[key];
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
              [key]: {
                ...state.buffers[key],
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