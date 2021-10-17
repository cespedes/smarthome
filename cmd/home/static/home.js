$(function() {
  $(".slider").after("<div></div>");
  $(".slider").slider({
    min: 0,
    max: 100,
    slide: function(event, ui) {
      $(this).next().html("[value=" + ui.value + "]");
    }
  });
});
