
Vue.component('xPostDto', {
    template: `
  <div style="margin: 5px 0px; display: flex;">

    <div  class="left" style="margin-right: 10px;">
        <textarea rows="15" cols="80">{{item.content}} </textarea>
    </div>
    <div style="min-width: 500px" class="right">
        <div><span style="font-weight:900;" >INDEX:</span><span>{{item.id}} - {{item.updated}}</span></div>
        <div><span style="font-weight:900;" >URI:</span><span>{{item.uri}}</span></div>
        <div><span style="font-weight:900;" >HOST:</span><span>{{item.host}}</span></div>
        <div><span style="font-weight:900;" >PROTOCOL:</span><span>{{item.protocol}}</span></div>
        <div class="headers" style="">
            <span style="font-weight:900;" >HEADERS:</span>
            <div v-for="header in item.headers"> {{header}} </div>
        </div>
        <div><span style="font-weight:900;" >RESPONSE TYPE:</span><span>{{item.responseType}}</span></div>
    </div>

  </div>
  `,
    props: {
        item: Object
    },
    data() {
        return {}
    },
    methods: {},
})
