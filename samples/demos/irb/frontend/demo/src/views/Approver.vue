<template>
  <div>

    <ProposalList
        v-if="listProposals"
        v-on:select-review="onSelectReview"
    />
    <ProposalReview
        v-else
        :reviewProposalWithId="selectProposalId"
        v-on:close-review="onCloseReview"
    />

  </div>
</template>

<script>
// @ is an alias to /src
import ProposalList from '@/components/proposal/List.vue';
import ProposalReview from '@/components/proposal/Review.vue';

export default {
  name: 'Approver',
  components: {
    ProposalList,
    ProposalReview,
  },
  data: () => ({
    details: true,
    selectProposalId: '',
  }),
  methods: {
    onSelectReview: function (proposalId) {
      this.selectProposalId = proposalId;
    },
    onCloseReview: function () {
      this.selectProposalId = '';
    }
  },
  computed: {
    listProposals() {
      return this.selectProposalId === '';
    }
  },
};
</script>
