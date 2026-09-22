import {reactive}from'vue';import type{Repair,RepairPriority}from'../types';export const repairStore=reactive<{items:Repair[];priority:RepairPriority}>({items:[],priority:'normal'});
