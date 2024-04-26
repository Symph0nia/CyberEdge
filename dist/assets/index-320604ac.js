import{d as u,c as o,n as h}from"./index-8fffec42.js";import{C as r}from"./close-circle-filled-91d2905f.js";import{C as d,E as p}from"./error-circle-filled-526edfc8.js";const f={components:{SearchIcon:u},data(){return this.$createElement,{data:[],originalData:[],size:"medium",tableLayout:!1,stripe:!0,bordered:!0,hover:!1,showHeader:!0,selectedRowKeys:[1],searchData:"",isSearchFocus:!1,columns:[{colKey:"id",title:"#",width:"50",ellipsis:!0},{colKey:"url",title:"URL",width:"300",cell:(t,{row:e})=>t("t-link",{attrs:{theme:"primary",hover:"color",href:e.url,target:"_blank"}},[e.url,t("JumpIcon",{slot:"suffixIcon"})])},{colKey:"content_type",title:"内容类型",width:"200",ellipsis:!0},{colKey:"status",title:"状态码",width:"100",cell:(t,{row:e})=>{let a={label:`HTTP (${e.status})`,theme:"default",icon:t(r)};return e.status>=200&&e.status<300?a={label:`HTTP (${e.status})`,theme:"success",icon:t(d)}:e.status>=300&&e.status<500?a={label:`HTTP (${e.status})`,theme:"warning",icon:t(p)}:e.status>=500&&(a={label:`HTTP (${e.status})`,theme:"danger",icon:t(r)}),t("t-tag",{attrs:{shape:"round",theme:a.theme,variant:"light-outline"}},[a.icon,a.label])}},{colKey:"length",title:"响应长度",width:"150"},{colKey:"operation",title:"操作",width:120,cell:(t,{row:e})=>t("t-button",{attrs:{theme:"danger",ghost:!0},on:{click:()=>this.deleteResult(e.id)}},["删除"])}],pagination:{current:1,pageSize:50,total:0,showJumper:!0}}},methods:{changeSearchFocus(t){this.isSearchFocus=t,t||this.filterData()},filterData(){this.searchData?this.data=this.originalData.filter(t=>Object.values(t).some(a=>a&&a.toString().toLowerCase().includes(this.searchData.toLowerCase()))):this.data=this.originalData},fetchResults(t){const e={task_id:t};this.$request.post("/api/path_scanner/task_status",JSON.stringify(e),{headers:{"Content-Type":"application/json"}}).then(a=>{const n=a.data.task_result.results,s=n.length;s>0?(o.success(`路径扫描结果已获取，共${s}条数据。`),this.data=n,this.originalData=n,this.pagination.total=s):o.info("没有获取到路径扫描结果。")}).catch(a=>{console.log(a),o.error("获取失败")})},deleteResult(t){this.$request.delete(`/api/path_scanner/paths/${t}/delete`,{headers:{"Content-Type":"application/json"}}).then(()=>{o.success("结果删除成功"),this.data=this.data.filter(e=>e.id!==t),this.pagination.total=this.data.length}).catch(e=>{console.error(e),o.error("删除失败")})},rehandleSelectChange(t,{selectedRowData:e}){this.selectedRowKeys=t},onPageChange(t,e){this.pagination.defaultCurrent||(this.pagination.current=t.current,this.pagination.pageSize=t.pageSize)},exportToCSV(){const t=[],e=["ID","URL","内容类型","状态码","响应长度"];t.push(e.join(",")),this.data.forEach(i=>{const c=[i.id,`"${i.url}"`,i.content_type,i.status,i.length];t.push(c.join(","))});const a=t.join(`
`),n=new Blob([a],{type:"text/csv;charset=utf-8;"}),s=document.createElement("a");s.href=URL.createObjectURL(n),s.download="path-scan-results.csv",s.style.display="none",document.body.appendChild(s),s.click(),document.body.removeChild(s)}},mounted(){const t=this.$route.query.task_id;this.fetchResults(t)}};var g=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("t-space",{attrs:{direction:"vertical"}},[a("t-input",{class:{"hover-active":t.isSearchFocus},attrs:{placeholder:"请输入搜索内容"},on:{blur:function(n){return t.changeSearchFocus(!1)},focus:function(n){return t.changeSearchFocus(!0)},input:t.filterData},scopedSlots:t._u([{key:"prefix-icon",fn:function(){return[a("search-icon",{staticClass:"icon",attrs:{size:"16"}})]},proxy:!0}]),model:{value:t.searchData,callback:function(n){t.searchData=n},expression:"searchData"}}),a("t-button",{attrs:{theme:"primary"},on:{click:t.exportToCSV}},[t._v("导出CSV")]),a("t-table",{attrs:{rowKey:"id",data:t.data,columns:t.columns,stripe:t.stripe,bordered:t.bordered,hover:t.hover,size:t.size,"table-layout":t.tableLayout?"auto":"fixed",pagination:t.pagination,showHeader:t.showHeader,"selected-row-keys":t.selectedRowKeys,cellEmptyContent:"-",resizable:""},on:{"select-change":t.rehandleSelectChange,"page-change":t.onPageChange}})],1)},m=[];const l={};var _=h(f,g,m,!1,y,null,null,null);function y(t){for(let e in l)this[e]=l[e]}const S=function(){return _.exports}();export{S as default};