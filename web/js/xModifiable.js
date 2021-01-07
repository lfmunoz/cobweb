
Vue.component('xModifiable', {
    template: `
  <div style="margin: 5px 0px">
      <label style="font-size: 0.9rem; display: inline-block;text-align: right;font-weight:900;min-width:170px">{{label}}:</label>
      <input style="width: 300px;" :value="value" @keyup="$emit('input', $event.target.value)" :readonly="readonly"/>
  </div>
  `,
    props: {
        label: {
            type: String
        },
        value: {
            type: String
        },
        readonly: {
            type: Boolean
        },
        text: {
            type: String,
            default: "Post"
        }
    },
    data() {
        return {}
    },
    methods: {},
})
