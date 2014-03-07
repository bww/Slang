(function($,log,_,Backbone){
  
  window.AboutView = Backbone.View.extend({
    
    el: '.page-about',
    
    events: {
      'click .accordion-toggle' : 'toggleAccordion',
    },
    
    initialize: function() {
      log('Created AboutView');
    },
    
    toggleAccordion: function(e) {
      e.preventDefault();
      var clickedLink = $(e.currentTarget);
      var section = clickedLink.closest('.accordion-section');
      section.toggleClass('active');
    }
    
  });
  
}(jQuery,window.log,window._,window.Backbone));
