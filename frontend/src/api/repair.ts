import {request} from '../utils/request';import type{Repair,RepairStatus,RepairUrgency}from'../types';
export interface ListRepairsParams{status?:string;overdue?:boolean}
export const listRepairs=(params:ListRepairsParams={})=>{
  const q=new URLSearchParams();
  if(params.status)q.set('status',params.status);
  if(params.overdue)q.set('overdue','1');
  const s=q.toString();
  return request<Repair[]>(`/repairs${s?`?${s}`:''}`);
};
export const createRepair=(data:{title:string;description:string;type:string;urgency?:RepairUrgency;images?:string})=>request<Repair>('/repairs',{method:'POST',body:JSON.stringify(data)});
export const assignRepair=(id:number,handler_id:number)=>request<Repair>(`/repairs/${id}/assign`,{method:'PATCH',body:JSON.stringify({handler_id})});
export const updateRepairStatus=(id:number,status:RepairStatus)=>request<Repair>(`/repairs/${id}/status`,{method:'PATCH',body:JSON.stringify({status})});