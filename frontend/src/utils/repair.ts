import type {Repair} from '../types';
import {repairUrgencyHours} from '../constants/repair';

export const formatDateTime=(v?:string)=>v?new Date(v).toLocaleString():'';

// 首次响应耗时（秒）转中文时长
export function formatDuration(seconds?:number):string{
  if(seconds===undefined||seconds===null)return '';
  if(seconds<0)seconds=0;
  if(seconds<3600)return `${Math.floor(seconds/60)}分钟`;
  return `${Math.floor(seconds/3600)}小时${Math.floor((seconds%3600)/60)}分钟`;
}

// 距首次响应截止的剩余时间；已超时返回空串（超时由标签呈现）
export function formatCountdown(repair:Repair,now=Date.now()):string{
  if(!repair.response_deadline||repair.responded_at)return '';
  const ms=new Date(repair.response_deadline).getTime()-now;
  const abs=Math.abs(ms);
  const h=Math.floor(abs/3_600_000),m=Math.floor((abs%3_600_000)/60_000);
  const text=h>0?`${h}小时${m}分钟`:`${m}分钟`;
  return ms>0?`剩余 ${text}`:`已超时 ${text}`;
}

export {repairUrgencyHours};
