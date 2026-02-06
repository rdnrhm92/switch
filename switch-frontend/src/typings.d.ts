declare module 'slash2';
declare module '*.css';
declare module '*.less';
declare module '*.scss';
declare module '*.sass';
declare module '*.svg';
declare module '*.png';
declare module '*.jpg';
declare module '*.jpeg';
declare module '*.gif';
declare module '*.bmp';
declare module '*.tiff';
declare module 'omit.js';
declare module 'numeral';
declare module 'mockjs';
declare module 'react-fittext';

declare const API_URL: string;


interface JumpData<T = any> {
  flag: string
  data: T
}

type Select = {
  status: string;
  id: number;
  name: string;
  tag: string
}

interface CommonModel {
  id: number;
  createdBy: string;
  updateBy: string;
  createTime: string | null;
  updateTime: string | null;
}


